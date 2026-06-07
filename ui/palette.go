package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Color literals live here once so the palette, the status helpers, and the spinner can't drift
// (roadmap M7 refinement). The three semantic colors are fixed — they read on both backgrounds —
// and also surface as Palette.Success/Warn/Error. The status helpers are fixed-color by API (their
// io.Writer signature carries no light/dark context): they can't resolve an adaptive pair, so Info
// uses the light (darker) variant of the palette's muted neutral, which stays legible on either bg.
var (
	colorSuccess = lipgloss.Color("#22C55E") // ✓, quoted strings, Palette.Success
	colorWarn    = lipgloss.Color("#F59E0B") // !, Palette.Warn
	colorError   = lipgloss.Color("#EF4444") // ✗, error details, Palette.Error
	colorSpinner = lipgloss.Color("#FFAF00") // animated spinner glyph (256-color 214, as hex)

	// Muted neutral pair; Palette.Muted resolves it by background. Info, being fixed-color, takes
	// the light/darker variant by reference so the two can't drift.
	mutedLight = lipgloss.Color("#6E7781")
	mutedDark  = lipgloss.Color("#8B949E")
	colorInfo  = mutedLight
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
		Muted:    ld(mutedLight, mutedDark),
		Dim:      adapt("#8C959F", "#6E7681"),
		Argument: adapt("#0969DA", "#79C0FF"),
		Success:  colorSuccess,
		Warn:     colorWarn,
		Error:    colorError,
	}
}
