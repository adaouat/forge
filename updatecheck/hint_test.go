package updatecheck

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixedNow() time.Time { return time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC) }

func TestHinter_NewerAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.3.0"}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	Hinter{Repo: "adaouat/heraut", Bin: "heraut", Current: "v1.2.0", BaseURL: srv.URL, Now: fixedNow}.
		Print(context.Background(), &buf)

	assert.Contains(t, buf.String(), "heraut v1.3.0 available")
	assert.Contains(t, buf.String(), "what's new: https://github.com/adaouat/heraut/releases/latest")
}

func TestUpgradeLine(t *testing.T) {
	const releases = "https://github.com/adaouat/heraut/releases/latest"
	tests := []struct {
		name string
		cmd  string
		want string
	}{
		{
			name: "package-manager command shows run and what's-new",
			cmd:  "brew upgrade heraut",
			want: "heraut v1.3.0 available — run: brew upgrade heraut · what's new: " + releases,
		},
		{
			name: "unknown install shows only the what's-new pointer",
			cmd:  "",
			want: "heraut v1.3.0 available · what's new: " + releases,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, upgradeLine("heraut", "v1.3.0", tc.cmd, releases))
		})
	}
}

func TestHinter_UpToDate_NoOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.0"}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	Hinter{Repo: "adaouat/heraut", Bin: "heraut", Current: "v1.2.0", BaseURL: srv.URL, Now: fixedNow}.
		Print(context.Background(), &buf)
	assert.Empty(t, buf.String())
}

func TestHinter_ErrorSwallowed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	require.NotPanics(t, func() {
		Hinter{Repo: "adaouat/heraut", Bin: "heraut", Current: "v1.2.0", BaseURL: srv.URL, Now: fixedNow}.
			Print(context.Background(), &buf)
	})
	assert.Empty(t, buf.String(), "errors are swallowed, nothing printed")
}

func TestHinter_FreshCacheSkipsFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("a fresh cache must not trigger a fetch")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cacheFile := filepath.Join(t.TempDir(), "check.json")
	data, _ := json.Marshal(cacheEntry{CheckedAt: fixedNow().Add(-1 * time.Hour), Latest: "v1.3.0"})
	require.NoError(t, os.WriteFile(cacheFile, data, 0o600))

	var buf bytes.Buffer
	Hinter{Repo: "adaouat/heraut", Bin: "heraut", Current: "v1.2.0", BaseURL: srv.URL, CacheFile: cacheFile, Now: fixedNow}.
		Print(context.Background(), &buf)
	assert.Contains(t, buf.String(), "heraut v1.3.0 available")
}

func TestHinter_CachesReleaseBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.3.0","body":"## What changed","html_url":"https://x/v1.3.0"}`))
	}))
	defer srv.Close()

	cacheFile := filepath.Join(t.TempDir(), "check.json")
	var buf bytes.Buffer
	Hinter{Repo: "adaouat/heraut", Bin: "heraut", Current: "v1.2.0", BaseURL: srv.URL, CacheFile: cacheFile, Now: fixedNow}.
		Print(context.Background(), &buf)

	data, err := os.ReadFile(cacheFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "What changed", "release body is cached for whatsnew's offline fallback")
	assert.Contains(t, string(data), "https://x/v1.3.0", "release url is cached")
}

func TestHinter_StaleCacheRefetches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v2.0.0"}`))
	}))
	defer srv.Close()

	cacheFile := filepath.Join(t.TempDir(), "check.json")
	data, _ := json.Marshal(cacheEntry{CheckedAt: fixedNow().Add(-48 * time.Hour), Latest: "v1.3.0"})
	require.NoError(t, os.WriteFile(cacheFile, data, 0o600))

	var buf bytes.Buffer
	Hinter{Repo: "adaouat/heraut", Bin: "heraut", Current: "v1.2.0", BaseURL: srv.URL, CacheFile: cacheFile, Now: fixedNow}.
		Print(context.Background(), &buf)

	assert.Contains(t, buf.String(), "v2.0.0", "stale cache should refetch")
	updated, _ := os.ReadFile(cacheFile)
	assert.Contains(t, string(updated), "v2.0.0", "cache is refreshed")
}
