package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStripANSI_stripsAndRestoresWholeTree(t *testing.T) {
	root := &cobra.Command{
		Use:   "root",
		Short: "\x1b[31mroot short\x1b[0m",
		Long:  "\x1b[31mroot long\x1b[0m",
	}
	child := &cobra.Command{
		Use:     "child",
		Short:   "\x1b[32mchild short\x1b[0m",
		Example: "\x1b[32mchild example\x1b[0m",
	}
	root.AddCommand(child)

	restore := stripANSI(root)

	assert.Equal(t, "root short", root.Short)
	assert.Equal(t, "root long", root.Long)
	assert.Equal(t, "child short", child.Short)
	assert.Equal(t, "child example", child.Example)

	restore()

	assert.Equal(t, "\x1b[31mroot short\x1b[0m", root.Short)
	assert.Equal(t, "\x1b[31mroot long\x1b[0m", root.Long)
	assert.Equal(t, "\x1b[32mchild short\x1b[0m", child.Short)
	assert.Equal(t, "\x1b[32mchild example\x1b[0m", child.Example)
}

func TestManCommand_RunE_emitsANSIFreeManPage(t *testing.T) {
	const styledLong = "\x1b[1;38;2;227;179;65mBANNER\x1b[m\ncatchphrase"

	root := &cobra.Command{Use: "tool", Short: "does things", Long: styledLong}
	root.AddCommand(manCommand())

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"man"})

	err := root.Execute()
	require.NoError(t, err)

	assert.NotContains(t, out.String(), "\x1b[", "man page must not contain raw ANSI escape codes")
	assert.Contains(t, out.String(), "BANNER")
	assert.Contains(t, out.String(), "catchphrase")
	assert.Equal(t, styledLong, root.Long, "root.Long must be restored after generating the man page")
}
