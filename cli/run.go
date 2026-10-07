// Package cli wraps the family's CLI framework (fang) so tools run through forge — the fang
// version, the theme, and signal handling live here, not in each tool. See forge ADR-0010.
package cli

import (
	"context"
	"os"
	"syscall"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/adaouat/forge/ui"
)

// Run executes cmd with fang, wiring the build version and the family theme (the shared palette
// plus the tool's accent). cmd.Context() is canceled on SIGINT/SIGTERM. fang's built-in `man`
// page generator is replaced with an ANSI-safe one (see man.go) so a tool's styled Long/Short
// (e.g. a colored banner) never leaks raw escape codes into the generated roff. Tools call this
// instead of fang.Execute.
func Run(ctx context.Context, cmd *cobra.Command, version string, accent ui.Accent) error {
	cmd.AddCommand(manCommand())
	return fang.Execute(ctx, cmd,
		fang.WithVersion(version),
		fang.WithColorSchemeFunc(func(ld lipgloss.LightDarkFunc) fang.ColorScheme {
			return ui.ColorScheme(ld, accent)
		}),
		fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM),
		fang.WithoutManpage(),
		fang.WithErrorHandler(errorHandler),
	)
}
