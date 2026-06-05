package ui

import (
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// HuhTheme builds a huh form theme branded with the shared palette plus a tool's accent (a zero
// Accent uses the Ember default), so interactive prompts match the tool's fang theme. Pass it to
// form.WithTheme. It starts from huh.ThemeBase and overrides the focused state + errors. See
// forge ADR-0010.
func HuhTheme(a Accent) huh.ThemeFunc {
	return func(isDark bool) *huh.Styles {
		s := huh.ThemeBase(isDark)
		ld := lipgloss.LightDark(isDark)
		p := NewPalette(ld)
		a = a.orDefault()
		accent := ld(a.Light, a.Dark)

		s.Focused.Title = s.Focused.Title.Foreground(accent)
		s.Focused.NoteTitle = s.Focused.NoteTitle.Foreground(accent)
		s.Focused.SelectSelector = s.Focused.SelectSelector.Foreground(accent)
		s.Focused.MultiSelectSelector = s.Focused.MultiSelectSelector.Foreground(accent)
		s.Focused.SelectedOption = s.Focused.SelectedOption.Foreground(accent)
		s.Focused.SelectedPrefix = s.Focused.SelectedPrefix.Foreground(accent)
		s.Focused.FocusedButton = s.Focused.FocusedButton.Background(accent).Foreground(lipgloss.Color("#FFFFFF"))
		s.Focused.Description = s.Focused.Description.Foreground(p.Muted)
		s.Focused.ErrorIndicator = s.Focused.ErrorIndicator.Foreground(p.Error)
		s.Focused.ErrorMessage = s.Focused.ErrorMessage.Foreground(p.Error)

		return s
	}
}
