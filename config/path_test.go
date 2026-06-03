package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/forge/config"
)

func TestResolve_FlagWins(t *testing.T) {
	t.Setenv("HERAUT_FILE", "/env/heraut.yml")
	path, src := config.Resolver{App: "heraut"}.Resolve("/flag/heraut.yml")
	assert.Equal(t, "/flag/heraut.yml", path)
	assert.Equal(t, config.FromFlag, src)
}

func TestResolve_EnvWinsOverDiscovery(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HERAUT_FILE", "/env/heraut.yml")
	path, src := config.Resolver{App: "heraut"}.Resolve("")
	assert.Equal(t, "/env/heraut.yml", path)
	assert.Equal(t, config.FromEnv, src)
}

func TestResolve_WhitespaceEnvFallsThrough(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HERAUT_FILE", "   ")
	path, src := config.Resolver{App: "heraut"}.Resolve("")
	assert.Equal(t, ".heraut.yml", path)
	assert.Equal(t, config.FromDefault, src)
}

func TestResolve_XDGWhenPresent(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HERAUT_FILE", "")
	require.NoError(t, os.MkdirAll(".config", 0o755))
	require.NoError(t, os.WriteFile(".config/heraut.yml", []byte("x: 1\n"), 0o600))
	path, src := config.Resolver{App: "heraut"}.Resolve("")
	assert.Equal(t, ".config/heraut.yml", path)
	assert.Equal(t, config.FromXDG, src)
}

func TestResolve_DefaultFallback(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HERAUT_FILE", "")
	path, src := config.Resolver{App: "heraut"}.Resolve("")
	assert.Equal(t, ".heraut.yml", path)
	assert.Equal(t, config.FromDefault, src)
}

func TestResolve_AppNameParameterized(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("BIFROST_FILE", "/env/bifrost.yml")
	path, src := config.Resolver{App: "bifrost"}.Resolve("")
	assert.Equal(t, "/env/bifrost.yml", path)
	assert.Equal(t, config.FromEnv, src)
}

func TestLabel(t *testing.T) {
	r := config.Resolver{App: "heraut"}
	assert.Equal(t, "--config", r.Label(config.FromFlag))
	assert.Equal(t, "HERAUT_FILE", r.Label(config.FromEnv))
	assert.Equal(t, ".config/heraut.yml", r.Label(config.FromXDG))
	assert.Equal(t, ".heraut.yml", r.Label(config.FromDefault))
}

func TestInitDest_XDGWhenConfigDirExists(t *testing.T) {
	t.Chdir(t.TempDir())
	require.NoError(t, os.MkdirAll(".config", 0o755))
	assert.Equal(t, ".config/heraut.yml", config.Resolver{App: "heraut"}.InitDest())
}

func TestInitDest_DefaultWhenNoConfigDir(t *testing.T) {
	t.Chdir(t.TempDir())
	assert.Equal(t, ".heraut.yml", config.Resolver{App: "heraut"}.InitDest())
}
