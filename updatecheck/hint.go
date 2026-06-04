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

	latest, ok := h.readCache(now())
	if !ok {
		l, err := Checker{Repo: h.Repo, BaseURL: h.BaseURL, Client: h.Client}.latest(ctx)
		if err != nil {
			return
		}
		latest = l
		h.writeCache(now(), latest)
	}

	if !isNewer(latest, h.Current) {
		return
	}
	cmd := DetectInstall().UpgradeCommand(h.Bin, h.Module)
	if cmd == "" {
		cmd = "see https://github.com/" + h.Repo + "/releases/latest"
	}
	_, _ = fmt.Fprintf(w, "%s %s available — run: %s\n", h.Bin, latest, cmd)
}

type cacheEntry struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
}

// readCache returns the cached latest tag if the cache is present and fresh.
func (h Hinter) readCache(now time.Time) (string, bool) {
	if h.CacheFile == "" {
		return "", false
	}
	data, err := os.ReadFile(h.CacheFile)
	if err != nil {
		return "", false
	}
	var e cacheEntry
	if err := json.Unmarshal(data, &e); err != nil {
		return "", false
	}
	if now.Sub(e.CheckedAt) >= cacheWindow {
		return "", false
	}
	return e.Latest, true
}

func (h Hinter) writeCache(now time.Time, latest string) {
	if h.CacheFile == "" {
		return
	}
	data, err := json.Marshal(cacheEntry{CheckedAt: now, Latest: latest})
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(h.CacheFile), 0o755)
	_ = os.WriteFile(h.CacheFile, data, 0o600)
}
