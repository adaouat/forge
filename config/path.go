package config

import (
	"os"
	"strings"
)

// Source identifies how a config file path was resolved.
type Source int

const (
	FromFlag    Source = iota // explicit --config flag
	FromEnv                   // <APP>_FILE environment variable
	FromXDG                   // .config/<app>.yml (auto-discovered)
	FromDefault               // .<app>.yml (fallback)
)

// Resolver resolves an app's config file path from the standard precedence.
// App is the lowercase CLI name, e.g. "heraut" → HERAUT_FILE, .config/heraut.yml,
// .heraut.yml.
type Resolver struct {
	App string
}

func (r Resolver) envVar() string      { return strings.ToUpper(r.App) + "_FILE" }
func (r Resolver) xdgPath() string     { return ".config/" + r.App + ".yml" }
func (r Resolver) defaultPath() string { return "." + r.App + ".yml" }

// Resolve returns the config path to use and the source that determined it:
//  1. explicit (--config) if non-empty → FromFlag
//  2. <APP>_FILE env var if set → FromEnv
//  3. .config/<app>.yml if it exists → FromXDG
//  4. .<app>.yml fallback → FromDefault
func (r Resolver) Resolve(explicit string) (string, Source) {
	if explicit != "" {
		return explicit, FromFlag
	}
	if env := strings.TrimSpace(os.Getenv(r.envVar())); env != "" {
		return env, FromEnv
	}
	if _, err := os.Stat(r.xdgPath()); err == nil {
		return r.xdgPath(), FromXDG
	}
	return r.defaultPath(), FromDefault
}

// Label returns a human-readable description of a Source, for messages like
// "config loaded from X (from <label>)".
func (r Resolver) Label(src Source) string {
	switch src {
	case FromFlag:
		return "--config"
	case FromEnv:
		return r.envVar()
	case FromXDG:
		return r.xdgPath()
	default:
		return r.defaultPath()
	}
}

// InitDest returns where a new config should be written: .config/<app>.yml if a
// .config/ directory exists, else .<app>.yml.
func (r Resolver) InitDest() string {
	if _, err := os.Stat(".config"); err == nil {
		return r.xdgPath()
	}
	return r.defaultPath()
}
