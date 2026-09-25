package cli

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
	mango "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
)

// manCommand replaces fang's built-in man page generator (disabled via
// fang.WithoutManpage in Run). fang's own RunE feeds the command tree straight into
// mango-cobra with no ANSI handling, so a tool that styles its Long/Short/Example (e.g. a
// colored banner) leaks raw escape codes into the roff output. This strips them first.
func manCommand() *cobra.Command {
	return &cobra.Command{
		Use:                   "man",
		Short:                 "Generates manpages",
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Args:                  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restore := stripANSI(cmd.Root())
			defer restore()

			page, err := mango.NewManPage(1, cmd.Root())
			if err != nil {
				return fmt.Errorf("building man page: %w", err)
			}
			if _, err := fmt.Fprint(cmd.OutOrStdout(), page.Build(roff.NewDocument())); err != nil {
				return fmt.Errorf("writing man page: %w", err)
			}
			return nil
		},
	}
}

// stripANSI replaces Short/Long/Example across root and every descendant with their
// ANSI-stripped form, returning a func that restores the originals.
func stripANSI(root *cobra.Command) func() {
	type saved struct {
		cmd                  *cobra.Command
		short, long, example string
	}

	var all []saved
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		all = append(all, saved{c, c.Short, c.Long, c.Example})
		c.Short = ansi.Strip(c.Short)
		c.Long = ansi.Strip(c.Long)
		c.Example = ansi.Strip(c.Example)
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)

	return func() {
		for _, s := range all {
			s.cmd.Short, s.cmd.Long, s.cmd.Example = s.short, s.long, s.example
		}
	}
}
