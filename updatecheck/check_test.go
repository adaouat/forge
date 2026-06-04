package updatecheck

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestCheckNewer_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, _, err := Checker{Repo: "adaouat/heraut", BaseURL: srv.URL}.CheckNewer(context.Background(), "v1.2.0")
	require.Error(t, err)
}
