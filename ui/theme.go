package ui

import (
	"image/color"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
)

// Accent is a tool's brand colors over the shared palette — the primary accent (titles, program
// name, flags) and the secondary (subcommands), each as a light/dark pair. It is the only
// per-tool part of the theme; everything else comes from the palette. See forge ADR-0010.
type Accent struct {
	Light, Dark                   color.Color // primary
	SecondaryLight, SecondaryDark color.Color // secondary
}

// DefaultAccent is forge's brand — "Ember", the forge fire (orange over coal-red). A tool that
// passes a zero Accent gets this; tools normally override it with their own.
func DefaultAccent() Accent {
	return Accent{
		Light:          lipgloss.Color("#EA580C"),
		Dark:           lipgloss.Color("#FB923C"),
		SecondaryLight: lipgloss.Color("#BE123C"),
		SecondaryDark:  lipgloss.Color("#FB7185"),
	}
}

func (a Accent) orDefault() Accent {
	if a.Light == nil && a.Dark == nil && a.SecondaryLight == nil && a.SecondaryDark == nil {
		return DefaultAccent()
	}
	return a
}

// ColorScheme builds a fang.ColorScheme from the shared palette plus a tool's accent (a zero
// Accent uses DefaultAccent). Apps pass this (via cli.Run) so the slot mapping lives once in
// forge, not duplicated per tool.
func ColorScheme(ld lipgloss.LightDarkFunc, a Accent) fang.ColorScheme {
	a = a.orDefault()
	p := NewPalette(ld)
	accent := ld(a.Light, a.Dark)
	secondary := ld(a.SecondaryLight, a.SecondaryDark)
	return fang.ColorScheme{
		Base:           p.Text,
		Title:          accent,
		Description:    p.Muted,
		Codeblock:      p.Muted,
		Program:        accent,
		DimmedArgument: p.Dim,
		Comment:        p.Muted,
		Flag:           accent,
		FlagDefault:    p.Dim,
		Command:        secondary,
		QuotedString:   p.Success,
		Argument:       p.Argument,
		Help:           p.Muted,
		Dash:           p.Muted,
		ErrorHeader:    [2]color.Color{lipgloss.Color("#FFFFFF"), p.Error},
		ErrorDetails:   p.Error,
	}
}
