package ui_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/forge/ui"
)

// All tests use *bytes.Buffer (not a TTY), so no spinner animates — only the
// resolved status line is written. This is the CI/pipe path.

func okFn() (ui.Result, error) { return ui.Result{}, nil }

func TestSpinner_Success_NoDetail(t *testing.T) {
	var w bytes.Buffer
	err := ui.NewSpinner(&w, ui.Human).Run("git", okFn)
	require.NoError(t, err)
	assert.Equal(t, "✓ git\n", w.String())
}

func TestSpinner_Success_WithDetail(t *testing.T) {
	var w bytes.Buffer
	err := ui.NewSpinner(&w, ui.Human).Run("git", func() (ui.Result, error) {
		return ui.Result{Detail: "2.49.0"}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, "✓ git — 2.49.0\n", w.String())
}

func TestSpinner_Fail_ReturnsErr(t *testing.T) {
	var w bytes.Buffer
	wantErr := errors.New("not configured; run: git config user.name <name>")
	err := ui.NewSpinner(&w, ui.Human).Run("git user.name", func() (ui.Result, error) {
		return ui.Result{}, wantErr
	})
	require.ErrorIs(t, err, wantErr)
	assert.Equal(t, "✗ git user.name — not configured; run: git config user.name <name>\n", w.String())
}

func TestSpinner_Fail_MultiLine(t *testing.T) {
	var w bytes.Buffer
	err := ui.NewSpinner(&w, ui.Human).Run("git", func() (ui.Result, error) {
		return ui.Result{}, errors.New("main error\n  hint: fix this")
	})
	require.Error(t, err)
	got := w.String()
	assert.True(t, strings.HasPrefix(got, "✗ git — main error\n"), got)
	assert.Contains(t, got, "hint: fix this")
}

func TestSpinner_Skip_IsAdvisory(t *testing.T) {
	var w bytes.Buffer
	err := ui.NewSpinner(&w, ui.Human).Run("working tree", func() (ui.Result, error) {
		return ui.Result{}, ui.Skip("2 uncommitted change(s)")
	})
	require.NoError(t, err, "skip is advisory, not a failure")
	assert.Equal(t, "! working tree — 2 uncommitted change(s)\n", w.String())
}

func TestSpinner_SubResults(t *testing.T) {
	var w bytes.Buffer
	err := ui.NewSpinner(&w, ui.Human).Run("release", func() (ui.Result, error) {
		return ui.Result{Detail: "v1.2.3", Subs: []string{"tag pushed", "assets uploaded"}}, nil
	})
	require.NoError(t, err)
	got := w.String()
	assert.Contains(t, got, "✓ release — v1.2.3\n")
	assert.Contains(t, got, "     ✓ tag pushed", "sub-results indented under the parent")
	assert.Contains(t, got, "     ✓ assets uploaded")
}

func TestSpinner_TotalCounter(t *testing.T) {
	var w bytes.Buffer
	sp := ui.NewSpinner(&w, ui.Human).Total(3)
	require.NoError(t, sp.Run("first", okFn))
	require.NoError(t, sp.Run("second", okFn))
	got := w.String()
	assert.Contains(t, got, "✓ [1/3] first\n")
	assert.Contains(t, got, "✓ [2/3] second\n")
}

func TestSpinner_NoCounterWhenTotalZero(t *testing.T) {
	var w bytes.Buffer
	require.NoError(t, ui.NewSpinner(&w, ui.Human).Run("solo", okFn))
	assert.Equal(t, "✓ solo\n", w.String(), "no [N/M] prefix without Total")
}

func TestSpinner_PlainMode_RendersStatusLine(t *testing.T) {
	var w bytes.Buffer
	err := ui.NewSpinner(&w, ui.Plain).Run("step", func() (ui.Result, error) {
		return ui.Result{Detail: "done"}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, "✓ step — done\n", w.String())
}
