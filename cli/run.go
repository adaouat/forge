// Package cli wraps the family's CLI framework (fang) so tools run through forge — the fang
// version and the theme live here, not in each tool. See forge ADR-0010.
package cli

import (
	"context"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/adaouat/forge/ui"
)

// Run executes cmd with fang, wiring the build version and the family theme (the shared palette
// plus the tool's accent). Tools call this instead of fang.Execute.
func Run(ctx context.Context, cmd *cobra.Command, version string, accent ui.Accent) error {
	return fang.Execute(ctx, cmd,
		fang.WithVersion(version),
		fang.WithColorSchemeFunc(func(ld lipgloss.LightDarkFunc) fang.ColorScheme {
			return ui.ColorScheme(ld, accent)
		}),
	)
}
