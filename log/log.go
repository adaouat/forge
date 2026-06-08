// Package log provides the family's logging foundation: a [log/slog] logger rendered through
// charm.land/log/v2, the same interface/implementation split as cli.Run/fang. See forge ADR-0011.
package log

import (
	"io"
	"log/slog"

	charmlog "charm.land/log/v2"
)

// New returns a [*slog.Logger] that renders leveled, colored, TTY-aware output through
// charm.land/log/v2 (which detects w's color profile itself). level sets the minimum level
// reported; w is typically os.Stderr.
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(charmlog.NewWithOptions(w, charmlog.Options{
		Level: charmlog.Level(level),
	}))
}
