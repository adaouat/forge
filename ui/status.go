package ui

import (
	"fmt"
	"io"

	"charm.land/lipgloss/v2"
)

// Success returns "✓ <msg>" styled green when w supports color, plain otherwise.
func Success(w io.Writer, msg string) string {
	if !HasColor(w) {
		return "✓ " + msg
	}
	return lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).Render("✓") + " " + msg
}

// Err returns "✗ <msg>" styled red when w supports color, plain otherwise.
func Err(w io.Writer, msg string) string {
	if !HasColor(w) {
		return "✗ " + msg
	}
	return lipgloss.NewStyle().Foreground(colorError).Bold(true).Render("✗") + " " + msg
}

// Warn returns "! <msg>" styled yellow when w supports color, plain otherwise.
// Use for skipped checks and non-critical notices.
func Warn(w io.Writer, msg string) string {
	if !HasColor(w) {
		return "! " + msg
	}
	return lipgloss.NewStyle().Foreground(colorWarn).Bold(true).Render("!") + " " + msg
}

// Info returns "  <msg>" dimmed when w supports color, plain otherwise.
// Use for hints and supplementary context below a primary status line.
func Info(w io.Writer, msg string) string {
	if !HasColor(w) {
		return "  " + msg
	}
	return lipgloss.NewStyle().Foreground(colorInfo).Render("  " + msg)
}

// Header writes a bold section title to w, surrounded by blank lines.
func Header(w io.Writer, title string) {
	if !HasColor(w) {
		_, _ = fmt.Fprintf(w, "\n%s\n\n", title)
		return
	}
	_, _ = fmt.Fprintf(w, "\n%s\n\n", lipgloss.NewStyle().Bold(true).Render(title))
}
