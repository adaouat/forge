package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const cacheWindow = 24 * time.Hour

// Hinter prints a one-line upgrade hint when a newer release exists, fetching at
// most once per 24h (when CacheFile is set). Every error is swallowed — the hint
// must never break a normal run.
type Hinter struct {
	Repo      string           // "owner/name"
	Bin       string           // binary name: the hint label and brew/mise/scoop command
	Module    string           // `go install` target path (used only when go-installed)
	Current   string           // current version
	CacheFile string           // "" → always fetch (no cache)
	BaseURL   string           // "" → https://api.github.com (test seam)
	Client    *http.Client     // nil → default
	Now       func() time.Time // nil → time.Now (test seam)
}

// Print writes "<bin> X available — run: <upgrade command>" to w when a newer
// release exists. Errors are swallowed.
func (h Hinter) Print(ctx context.Context, w io.Writer) {
	now := time.Now
	if h.Now != nil {
		now = h.Now
	}

	entry, ok := readCache(h.CacheFile, now())
	latest := entry.Latest
	if !ok {
		rel, err := Checker{Repo: h.Repo, BaseURL: h.BaseURL, Client: h.Client}.latestRelease(ctx)
		if err != nil {
			return
		}
		latest = rel.Tag
		writeCache(h.CacheFile, cacheEntry{CheckedAt: now(), Latest: rel.Tag, Body: rel.Body, URL: rel.URL})
	}

	if !isNewer(latest, h.Current) {
		return
	}
	releases := "https://github.com/" + h.Repo + "/releases/latest"
	cmd := DetectInstall().UpgradeCommand(h.Bin, h.Module)
	_, _ = fmt.Fprintln(w, upgradeLine(h.Bin, latest, cmd, releases))
}

// upgradeLine formats the one-line hint. The what's-new pointer is always present; the
// "run:" clause appears only when an install method was detected (cmd != ""), otherwise the
// releases page doubles as both the changelog and the download. See forge ADR-0012 (tier A).
func upgradeLine(bin, latest, cmd, releases string) string {
	if cmd == "" {
		return fmt.Sprintf("%s %s available · what's new: %s", bin, latest, releases)
	}
	return fmt.Sprintf("%s %s available — run: %s · what's new: %s", bin, latest, cmd, releases)
}

type cacheEntry struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
	Body      string    `json:"body,omitempty"`
	URL       string    `json:"url,omitempty"`
}

// readCache returns the cached entry if the cache file is present and fresh.
func readCache(path string, now time.Time) (cacheEntry, bool) {
	if path == "" {
		return cacheEntry{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cacheEntry{}, false
	}
	var e cacheEntry
	if err := json.Unmarshal(data, &e); err != nil {
		return cacheEntry{}, false
	}
	if now.Sub(e.CheckedAt) >= cacheWindow {
		return cacheEntry{}, false
	}
	return e, true
}

func writeCache(path string, e cacheEntry) {
	if path == "" {
		return
	}
	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, data, 0o600)
}
