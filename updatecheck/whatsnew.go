package updatecheck

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"charm.land/glamour/v2"
	"github.com/spf13/cobra"
)

// WhatsNewConfig parameterises [WhatsNewCommand] for one tool. Repo and Current are required;
// CacheFile enables the offline fallback, and BaseURL/Client/Now are test seams.
type WhatsNewConfig struct {
	Repo      string           // "owner/name"
	Current   string           // running version
	CacheFile string           // offline fallback source ("" → no fallback)
	BaseURL   string           // "" → https://api.github.com
	Client    *http.Client     // nil → default
	Now       func() time.Time // nil → time.Now
}

// WhatsNewCommand returns a `whatsnew` cobra command that renders the release notes for every
// version newer than the running build, glamour-rendered. See forge ADR-0012.
func WhatsNewCommand(cfg WhatsNewConfig) *cobra.Command {
	return &cobra.Command{
		Use:          "whatsnew",
		Short:        "Show release notes for versions newer than the running build",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cfg.run(cmd.Context(), cmd.OutOrStdout())
		},
	}
}

func (cfg WhatsNewConfig) run(ctx context.Context, w io.Writer) error {
	rels, err := Checker{Repo: cfg.Repo, BaseURL: cfg.BaseURL, Client: cfg.Client}.listReleases(ctx)
	if err != nil {
		if cached, ok := cfg.cachedNewer(); ok {
			return render(w, assemble(cached))
		}
		return fmt.Errorf("fetching releases: %w", err)
	}

	var newer []release
	for _, r := range rels {
		if isNewer(r.Tag, cfg.Current) {
			newer = append(newer, r)
		}
	}
	if len(newer) == 0 {
		_, err := fmt.Fprintf(w, "You're on the latest release (%s).\n", cfg.Current)
		return err
	}
	return render(w, assemble(newer))
}

// cachedNewer returns the cached latest release as a one-element span when the cache is fresh,
// has a body, and is newer than current — whatsnew's offline fallback.
func (cfg WhatsNewConfig) cachedNewer() ([]release, bool) {
	now := time.Now
	if cfg.Now != nil {
		now = cfg.Now
	}
	entry, ok := readCache(cfg.CacheFile, now())
	if !ok || entry.Body == "" || !isNewer(entry.Latest, cfg.Current) {
		return nil, false
	}
	return []release{{Tag: entry.Latest, Body: entry.Body, URL: entry.URL}}, true
}

// assemble renders the release list to a single markdown document, newest first.
func assemble(rels []release) string {
	var b strings.Builder
	for i, r := range rels {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "# %s\n\n%s\n", r.Tag, strings.TrimSpace(r.Body))
	}
	return b.String()
}

// render writes md to w through glamour, falling back to the raw markdown if glamour fails —
// the styled render is best-effort, but the content must always reach the user. See ADR-0012.
func render(w io.Writer, md string) error {
	out, err := glamour.Render(md, "auto")
	if err != nil {
		out = md
	}
	if _, err := io.WriteString(w, out); err != nil {
		return fmt.Errorf("writing changelog: %w", err)
	}
	return nil
}
