// Package ui provides shared terminal output helpers for the adaouat CLIs:
// color/TTY detection, status lines, and version/help banner renderers.
package ui

import (
	"io"
	"os"

	// colorprofile and x/term have no charm.land module path (their go.mod still
	// declares github.com/charmbracelet/*); documented exception to the registry
	// rule — see docs/rules/coding.md and ADR-0001 M3 flag.
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

// HasColor reports whether w supports ANSI color output. It delegates to
// colorprofile.Detect, which honours NO_COLOR, CLICOLOR_FORCE, TERM=dumb, and
// whether w is a terminal.
func HasColor(w io.Writer) bool {
	return colorprofile.Detect(w, os.Environ()) >= colorprofile.ANSI
}

// IsTTY reports whether w is a terminal.
func IsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(f.Fd())
}
