package config_test

import (
	"os"
	"path/filepath"
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

func TestExpandHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"bare tilde is home", "~", home},
		{"tilde slash is home", "~/", home},
		{"tilde path joins home", "~/.config/x.yml", filepath.Join(home, ".config/x.yml")},
		{"other user is not expanded", "~user/x", "~user/x"},
		{"tilde-prefixed name is not expanded", "~name.yml", "~name.yml"},
		{"tilde mid-path is not expanded", "/data/~/x", "/data/~/x"},
		{"absolute path unchanged", "/etc/x", "/etc/x"},
		{"relative path unchanged", "x/y", "x/y"},
		{"empty unchanged", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, config.ExpandHome(tc.in))
		})
	}
}

func TestExpandHome_NoHomeLeavesPathUnchanged(t *testing.T) {
	t.Setenv("HOME", "")
	assert.Equal(t, "~/x", config.ExpandHome("~/x"))
}

func TestResolve_ExpandsHome(t *testing.T) {
	home := t.TempDir()
	tests := []struct {
		name     string
		explicit string
		env      string
		want     string
		wantSrc  config.Source
	}{
		{"flag tilde path", "~/flag/heraut.yml", "", filepath.Join(home, "flag/heraut.yml"), config.FromFlag},
		{"env tilde path", "", "~/env/heraut.yml", filepath.Join(home, "env/heraut.yml"), config.FromEnv},
		{"env tilde path with whitespace", "", "  ~/env/heraut.yml  ", filepath.Join(home, "env/heraut.yml"), config.FromEnv},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("HOME", home)
			t.Setenv("HERAUT_FILE", tc.env)
			path, src := config.Resolver{App: "heraut"}.Resolve(tc.explicit)
			assert.Equal(t, tc.want, path)
			assert.Equal(t, tc.wantSrc, src)
		})
	}
}

func TestResolve_NoHomeLeavesTildeUnchanged(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HOME", "")
	t.Setenv("HERAUT_FILE", "~/env/heraut.yml")
	path, src := config.Resolver{App: "heraut"}.Resolve("")
	assert.Equal(t, "~/env/heraut.yml", path)
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
