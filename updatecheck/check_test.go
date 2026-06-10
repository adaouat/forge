package updatecheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckNewer_NewerAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/adaouat/heraut/releases/latest", r.URL.Path)
		_, _ = w.Write([]byte(`{"tag_name":"v1.3.0"}`))
	}))
	defer srv.Close()

	latest, newer, err := Checker{Repo: "adaouat/heraut", BaseURL: srv.URL}.CheckNewer(context.Background(), "v1.2.0")
	require.NoError(t, err)
	assert.Equal(t, "v1.3.0", latest)
	assert.True(t, newer)
}

func TestCheckNewer_UpToDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.0"}`))
	}))
	defer srv.Close()

	_, newer, err := Checker{Repo: "adaouat/heraut", BaseURL: srv.URL}.CheckNewer(context.Background(), "v1.2.0")
	require.NoError(t, err)
	assert.False(t, newer)
}

func TestCheckNewer_OversizedBodyCapped(t *testing.T) {
	// A body larger than the cap is truncated before decode, so the (now incomplete)
	// JSON fails to parse — proving the response body is bounded. Without the cap the
	// whole body would decode and no error would surface.
	huge := strings.Repeat("v", 1<<20)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"` + huge + `"}`))
	}))
	defer srv.Close()

	_, _, err := Checker{Repo: "adaouat/heraut", BaseURL: srv.URL}.CheckNewer(context.Background(), "v1.2.0")
	require.Error(t, err)
}

func TestCheckNewer_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, _, err := Checker{Repo: "adaouat/heraut", BaseURL: srv.URL}.CheckNewer(context.Background(), "v1.2.0")
	require.Error(t, err)
}

func TestChecker_latestRelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/adaouat/heraut/releases/latest", r.URL.Path)
		_, _ = w.Write([]byte(`{"tag_name":"v1.3.0","body":"## What changed\n- thing","html_url":"https://github.com/adaouat/heraut/releases/tag/v1.3.0"}`))
	}))
	defer srv.Close()

	rel, err := Checker{Repo: "adaouat/heraut", BaseURL: srv.URL}.latestRelease(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "v1.3.0", rel.Tag)
	assert.Equal(t, "## What changed\n- thing", rel.Body)
	assert.Equal(t, "https://github.com/adaouat/heraut/releases/tag/v1.3.0", rel.URL)
}

func TestChecker_listReleases(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/adaouat/heraut/releases", r.URL.Path)
		_, _ = w.Write([]byte(`[
			{"tag_name":"v1.3.0","body":"third","html_url":"u3"},
			{"tag_name":"v1.2.0","body":"second","html_url":"u2"}
		]`))
	}))
	defer srv.Close()

	rels, err := Checker{Repo: "adaouat/heraut", BaseURL: srv.URL}.listReleases(context.Background())
	require.NoError(t, err)
	require.Len(t, rels, 2)
	assert.Equal(t, "v1.3.0", rels[0].Tag)
	assert.Equal(t, "third", rels[0].Body)
	assert.Equal(t, "v1.2.0", rels[1].Tag)
}
