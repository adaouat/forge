package updatecheck

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeStub(t *testing.T, dir, name string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o755))
}

// unsetEnv removes key from the environment for the test, restoring its prior
// value (or leaving it unset) on cleanup. t.Setenv cannot represent "unset".
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, ok := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, old)
		}
	})
}

func TestResolvePager(t *testing.T) {
	dir := t.TempDir()
	writeStub(t, dir, "less")
	writeStub(t, dir, "cat")
	t.Setenv("PATH", dir)

	t.Run("NO_PAGER disables paging", func(t *testing.T) {
		t.Setenv("NO_PAGER", "1")
		t.Setenv("PAGER", "")
		_, _, ok := resolvePager()
		assert.False(t, ok)
	})

	t.Run("default less with LESS unset sets LESS=FRX", func(t *testing.T) {
		t.Setenv("NO_PAGER", "")
		t.Setenv("PAGER", "")
		unsetEnv(t, "LESS")
		path, extraEnv, ok := resolvePager()
		require.True(t, ok)
		assert.Equal(t, "less", filepath.Base(path))
		assert.Equal(t, []string{"LESS=FRX"}, extraEnv)
	})

	t.Run("default less with LESS already set is left alone", func(t *testing.T) {
		t.Setenv("NO_PAGER", "")
		t.Setenv("PAGER", "")
		t.Setenv("LESS", "-X")
		path, extraEnv, ok := resolvePager()
		require.True(t, ok)
		assert.Equal(t, "less", filepath.Base(path))
		assert.Empty(t, extraEnv)
	})

	t.Run("PAGER=cat is used as-is", func(t *testing.T) {
		t.Setenv("NO_PAGER", "")
		t.Setenv("PAGER", "cat")
		path, extraEnv, ok := resolvePager()
		require.True(t, ok)
		assert.Equal(t, "cat", filepath.Base(path))
		assert.Empty(t, extraEnv)
	})

	t.Run("unknown pager disables paging", func(t *testing.T) {
		t.Setenv("NO_PAGER", "")
		t.Setenv("PAGER", "nonexistent-pager")
		_, _, ok := resolvePager()
		assert.False(t, ok)
	})
}

func TestPagedOutput_NonTTY(t *testing.T) {
	var buf bytes.Buffer
	assert.False(t, pagedOutput(&buf, "content"))
	assert.Empty(t, buf.String(), "pagedOutput must not write to w itself")
}
