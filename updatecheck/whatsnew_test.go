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

func TestAssemble(t *testing.T) {
	tests := []struct {
		name string
		in   []release
		want string
	}{
		{
			name: "single release",
			in:   []release{{Tag: "v1.3.0", Body: "## What changed\n- thing"}},
			want: "# v1.3.0\n\n## What changed\n- thing\n",
		},
		{
			name: "span keeps newest-first order",
			in:   []release{{Tag: "v1.4.0", Body: "b4"}, {Tag: "v1.3.0", Body: "b3"}},
			want: "# v1.4.0\n\nb4\n\n# v1.3.0\n\nb3\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, assemble(tc.in))
		})
	}
}

func TestRender(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, render(&buf, "# Hello\n\nworld"))
	out := buf.String()
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Hello")
	assert.Contains(t, out, "world")
}

func TestGlamourStyle(t *testing.T) {
	t.Run("no color support -> notty", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		t.Setenv("CLICOLOR_FORCE", "")
		assert.Equal(t, "notty", glamourStyle(&bytes.Buffer{}))
	})

	t.Run("color forced -> dark", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		t.Setenv("CLICOLOR_FORCE", "1")
		t.Setenv("TERM", "xterm-256color")
		assert.Equal(t, "dark", glamourStyle(&bytes.Buffer{}))
	})
}

func TestWhatsNewCommand_RendersSpanNewerThanCurrent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/adaouat/heraut/releases", r.URL.Path)
		_, _ = w.Write([]byte(`[
			{"tag_name":"v1.4.0","body":"four notes"},
			{"tag_name":"v1.3.0","body":"three notes"},
			{"tag_name":"v1.2.0","body":"two notes"}
		]`))
	}))
	defer srv.Close()

	cmd := WhatsNewCommand(WhatsNewConfig{Repo: "adaouat/heraut", Current: "v1.2.0", BaseURL: srv.URL})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	out := buf.String()
	assert.Contains(t, out, "v1.4.0")
	assert.Contains(t, out, "four notes")
	assert.Contains(t, out, "three notes")
	assert.NotContains(t, out, "two notes", "v1.2.0 == current is excluded")
}

func TestWhatsNewCommand_UpToDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"tag_name":"v1.2.0","body":"two"}]`))
	}))
	defer srv.Close()

	cmd := WhatsNewCommand(WhatsNewConfig{Repo: "adaouat/heraut", Current: "v1.2.0", BaseURL: srv.URL})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, buf.String(), "latest release (v1.2.0)")
}

func TestWhatsNewConfig_OfflineFallsBackToCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cacheFile := filepath.Join(t.TempDir(), "check.json")
	data, _ := json.Marshal(cacheEntry{CheckedAt: fixedNow().Add(-time.Hour), Latest: "v1.3.0", Body: "cached notes", URL: "u"})
	require.NoError(t, os.WriteFile(cacheFile, data, 0o600))

	var buf bytes.Buffer
	cfg := WhatsNewConfig{Repo: "adaouat/heraut", Current: "v1.2.0", BaseURL: srv.URL, CacheFile: cacheFile, Now: fixedNow}
	require.NoError(t, cfg.run(context.Background(), &buf))
	assert.Contains(t, buf.String(), "cached notes")
}

func TestWhatsNewConfig_OfflineNoCacheErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	cfg := WhatsNewConfig{Repo: "adaouat/heraut", Current: "v1.2.0", BaseURL: srv.URL, Now: fixedNow}
	err := cfg.run(context.Background(), &buf)
	require.Error(t, err)
	assert.Empty(t, buf.String())
}

func TestWhatsNewConfig_OfflineFallsBackToEmbeddedChangelog(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	cfg := WhatsNewConfig{
		Repo: "adaouat/heraut", Current: "v1.2.0", BaseURL: srv.URL, Now: fixedNow,
		Changelog: "# Changelog\n\n## v1.2.0\n- baked into this build",
	}
	require.NoError(t, cfg.run(context.Background(), &buf))
	assert.Contains(t, buf.String(), "baked into this build")
}

func TestWhatsNewConfig_CacheBeatsEmbeddedOffline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cacheFile := filepath.Join(t.TempDir(), "check.json")
	data, _ := json.Marshal(cacheEntry{CheckedAt: fixedNow().Add(-time.Hour), Latest: "v1.3.0", Body: "cached notes", URL: "u"})
	require.NoError(t, os.WriteFile(cacheFile, data, 0o600))

	var buf bytes.Buffer
	cfg := WhatsNewConfig{
		Repo: "adaouat/heraut", Current: "v1.2.0", BaseURL: srv.URL, CacheFile: cacheFile, Now: fixedNow,
		Changelog: "embedded notes",
	}
	require.NoError(t, cfg.run(context.Background(), &buf))
	assert.Contains(t, buf.String(), "cached notes")
	assert.NotContains(t, buf.String(), "embedded notes", "fresh cache wins over the embedded fallback")
}
