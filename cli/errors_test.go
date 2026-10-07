package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/forge/ui"
)

func TestRun_errorTextKeepsIdentifiersVerbatim(t *testing.T) {
	a := ui.Accent{
		Light: lipgloss.Color("#0AAAAA"), Dark: lipgloss.Color("#0AAAAA"),
		SecondaryLight: lipgloss.Color("#AA00AA"), SecondaryDark: lipgloss.Color("#AA00AA"),
	}
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{"leading flag", "--pre-release cannot be combined with --set-version", "--pre-release cannot be combined with --set-version."},
		{"leading tag", "v1.4.0-beta.1 would sort below existing v1.4.0-rc.1", "v1.4.0-beta.1 would sort below existing v1.4.0-rc.1."},
		{"leading hyphenated word", "pre-release series v1.4.0-rc.2 would escalate", "pre-release series v1.4.0-rc.2 would escalate."},
		{"leading config key", "versioning.format: required", "versioning.format: required."},
		{"plain leading word is capitalised", "no commits since v1.0.0", "No commits since v1.0.0."},
		{"plain leading word with a trailing colon", "error: no config found", "Error: no config found."},
		{"already capitalised", "Manual bump mode requires --set-version", "Manual bump mode requires --set-version."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:  "x",
				RunE: func(*cobra.Command, []string) error { return errors.New(tc.msg) },
			}
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs([]string{})

			err := Run(context.Background(), cmd, "1.0.0", a)

			require.Error(t, err)
			rendered := strings.Join(strings.Fields(ansi.Strip(out.String())), " ")
			assert.Contains(t, rendered, tc.want)
		})
	}
}
