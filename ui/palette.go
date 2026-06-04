package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Palette is the family's shared structural colors — everything in a tool's theme except its
// per-tool accent. Resolve it with NewPalette, then an app builds its fang.ColorScheme from these
// plus its accent (see docs/adr/0008-ui-theme-palette.md). Semantic colors match the status
// helpers, so status output and the fang theme agree.
type Palette struct {
	Text     color.Color // primary text
	Muted    color.Color // descriptions, help, comments
	Dim      color.Color // flag defaults, dimmed arguments
	Argument color.Color // command arguments
	Success  color.Color // ✓ and quoted strings
	Warn     color.Color // !
	Error    color.Color // ✗ and error details
}

// NewPalette resolves the shared palette for the terminal background. Neutrals adapt to
// light/dark; the semantic colors are fixed (they read on both backgrounds and match status.go).
func NewPalette(ld lipgloss.LightDarkFunc) Palette {
	adapt := func(light, dark string) color.Color {
		return ld(lipgloss.Color(light), lipgloss.Color(dark))
	}
	return Palette{
		Text:     adapt("#24292F", "#E6EDF3"),
		Muted:    adapt("#6E7781", "#8B949E"),
		Dim:      adapt("#8C959F", "#6E7681"),
		Argument: adapt("#0969DA", "#79C0FF"),
		Success:  lipgloss.Color("#22C55E"),
		Warn:     lipgloss.Color("#F59E0B"),
		Error:    lipgloss.Color("#EF4444"),
	}
}
