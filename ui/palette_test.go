package ui

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewPalette(t *testing.T) {
	var pickLight lipgloss.LightDarkFunc = func(light, _ color.Color) color.Color { return light }
	var pickDark lipgloss.LightDarkFunc = func(_, dark color.Color) color.Color { return dark }

	tests := []struct {
		name string
		ld   lipgloss.LightDarkFunc
		want Palette
	}{
		{
			name: "dark background picks dark neutrals; semantic colors are fixed",
			ld:   pickDark,
			want: Palette{
				Text:     lipgloss.Color("#E6EDF3"),
				Muted:    lipgloss.Color("#8B949E"),
				Dim:      lipgloss.Color("#6E7681"),
				Argument: lipgloss.Color("#79C0FF"),
				Success:  lipgloss.Color("#22C55E"),
				Warn:     lipgloss.Color("#F59E0B"),
				Error:    lipgloss.Color("#EF4444"),
			},
		},
		{
			name: "light background picks light neutrals; semantic colors unchanged",
			ld:   pickLight,
			want: Palette{
				Text:     lipgloss.Color("#24292F"),
				Muted:    lipgloss.Color("#6E7781"),
				Dim:      lipgloss.Color("#8C959F"),
				Argument: lipgloss.Color("#0969DA"),
				Success:  lipgloss.Color("#22C55E"),
				Warn:     lipgloss.Color("#F59E0B"),
				Error:    lipgloss.Color("#EF4444"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, NewPalette(tc.ld))
		})
	}
}
