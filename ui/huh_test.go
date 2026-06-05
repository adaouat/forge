package ui

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

func TestHuhTheme(t *testing.T) {
	a := Accent{
		Light:          lipgloss.Color("#111111"),
		Dark:           lipgloss.Color("#AAAAAA"),
		SecondaryLight: lipgloss.Color("#222222"),
		SecondaryDark:  lipgloss.Color("#BBBBBB"),
	}
	s := HuhTheme(a)(true) // dark
	p := NewPalette(lipgloss.LightDark(true))

	assert.Equal(t, lipgloss.Color("#AAAAAA"), s.Focused.Title.GetForeground(), "focused title = accent")
	assert.Equal(t, lipgloss.Color("#AAAAAA"), s.Focused.SelectSelector.GetForeground())
	assert.Equal(t, lipgloss.Color("#AAAAAA"), s.Focused.SelectedOption.GetForeground())
	assert.Equal(t, p.Error, s.Focused.ErrorMessage.GetForeground(), "errors from the palette")
}

func TestHuhTheme_zeroAccentUsesEmberDefault(t *testing.T) {
	s := HuhTheme(Accent{})(true)
	assert.Equal(t, DefaultAccent().Dark, s.Focused.Title.GetForeground())
}
