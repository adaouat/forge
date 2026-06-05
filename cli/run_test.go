package cli

import (
	"context"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/adaouat/forge/ui"
)

func TestRun_executesCommand(t *testing.T) {
	a := ui.Accent{
		Light: lipgloss.Color("#0AAAAA"), Dark: lipgloss.Color("#0AAAAA"),
		SecondaryLight: lipgloss.Color("#AA00AA"), SecondaryDark: lipgloss.Color("#AA00AA"),
	}
	ran := false
	cmd := &cobra.Command{
		Use:  "x",
		RunE: func(*cobra.Command, []string) error { ran = true; return nil },
	}
	cmd.SetArgs([]string{})

	err := Run(context.Background(), cmd, "1.0.0", a)
	assert.NoError(t, err)
	assert.True(t, ran, "Run should execute the command's RunE")
}
