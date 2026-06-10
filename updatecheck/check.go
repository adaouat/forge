package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.github.com"

// maxResponseBytes caps the GitHub response body before decoding. The latest-release
// payload is a few KB; 1 MiB is generous headroom and bounds a hostile/buggy response.
const maxResponseBytes = 1 << 20

// Checker queries a GitHub repo's latest release.
type Checker struct {
	Repo    string       // "owner/name"
	BaseURL string       // "" → https://api.github.com
	Client  *http.Client // nil → a client with a 5s timeout
}

// CheckNewer returns the latest release tag for the repo and whether it is newer
// than current.
func (c Checker) CheckNewer(ctx context.Context, current string) (latest string, newer bool, err error) {
	rel, err := c.latestRelease(ctx)
	if err != nil {
		return "", false, err
	}
	return rel.Tag, isNewer(rel.Tag, current), nil
}

// release is a GitHub release: its tag, the release-note markdown, and the page URL.
type release struct {
	Tag  string `json:"tag_name"`
	Body string `json:"body"`
	URL  string `json:"html_url"`
}

func (c Checker) latestRelease(ctx context.Context) (release, error) {
	var rel release
	if err := c.fetchJSON(ctx, "releases/latest", &rel); err != nil {
		return release{}, err
	}
	if rel.Tag == "" {
		return release{}, fmt.Errorf("update check: empty tag_name")
	}
	return rel, nil
}

func (c Checker) listReleases(ctx context.Context) ([]release, error) {
	var rels []release
	if err := c.fetchJSON(ctx, "releases", &rels); err != nil {
		return nil, err
	}
	return rels, nil
}

func (c Checker) fetchJSON(ctx context.Context, path string, into any) error {
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	url := fmt.Sprintf("%s/repos/%s/%s", base, c.Repo, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("update check request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("update check: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update check: github returned %s", resp.Status)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(into); err != nil {
		return fmt.Errorf("update check decode: %w", err)
	}
	return nil
}
