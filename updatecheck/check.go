package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.github.com"

// Checker queries a GitHub repo's latest release.
type Checker struct {
	Repo    string       // "owner/name"
	BaseURL string       // "" → https://api.github.com
	Client  *http.Client // nil → a client with a 5s timeout
}

// CheckNewer returns the latest release tag for the repo and whether it is newer
// than current.
func (c Checker) CheckNewer(ctx context.Context, current string) (latest string, newer bool, err error) {
	latest, err = c.latest(ctx)
	if err != nil {
		return "", false, err
	}
	return latest, isNewer(latest, current), nil
}

func (c Checker) latest(ctx context.Context) (string, error) {
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	url := fmt.Sprintf("%s/repos/%s/releases/latest", base, c.Repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("update check request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("update check: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("update check: github returned %s", resp.Status)
	}

	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", fmt.Errorf("update check decode: %w", err)
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("update check: empty tag_name")
	}
	return rel.TagName, nil
}
