package ui

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

func TestColorScheme(t *testing.T) {
	var dark lipgloss.LightDarkFunc = func(_, d color.Color) color.Color { return d }

	a := Accent{
		Light:          lipgloss.Color("#111111"),
		Dark:           lipgloss.Color("#AAAAAA"),
		SecondaryLight: lipgloss.Color("#222222"),
		SecondaryDark:  lipgloss.Color("#BBBBBB"),
	}
	cs := ColorScheme(dark, a)
	p := NewPalette(dark)

	// accent drives the brand slots; secondary drives commands; the rest is the shared palette.
	assert.Equal(t, lipgloss.Color("#AAAAAA"), cs.Title, "title = accent (dark)")
	assert.Equal(t, lipgloss.Color("#AAAAAA"), cs.Program)
	assert.Equal(t, lipgloss.Color("#AAAAAA"), cs.Flag)
	assert.Equal(t, lipgloss.Color("#BBBBBB"), cs.Command, "command = secondary (dark)")
	assert.Equal(t, p.Text, cs.Base)
	assert.Equal(t, p.Argument, cs.Argument)
	assert.Equal(t, p.Success, cs.QuotedString)
	assert.Equal(t, p.Muted, cs.Description)
	assert.Equal(t, p.Dim, cs.FlagDefault)
	// Codeblock is the usage-block *background*, so it must be the subtle surface — not p.Muted, or
	// the DimmedArgument placeholders render gray-on-gray (the [command]/[--flags] visibility bug).
	assert.Equal(t, p.Surface, cs.Codeblock)
	assert.NotEqual(t, cs.Codeblock, cs.DimmedArgument, "usage placeholders must contrast with the block")
	assert.Equal(t, p.Error, cs.ErrorDetails)
	assert.Equal(t, p.Error, cs.ErrorHeader[1])
}

func TestColorScheme_zeroAccentUsesEmberDefault(t *testing.T) {
	var dark lipgloss.LightDarkFunc = func(_, d color.Color) color.Color { return d }

	cs := ColorScheme(dark, Accent{})
	def := DefaultAccent()
	assert.Equal(t, def.Dark, cs.Title, "a zero accent falls back to the Ember default")
	assert.Equal(t, def.SecondaryDark, cs.Command)
}
