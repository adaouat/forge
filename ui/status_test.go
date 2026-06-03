package ui_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adaouat/forge/ui"
)

// Most tests use *bytes.Buffer as the writer, which is not a TTY:
// colorprofile.Detect returns NoTTY for a buffer → helpers return plain text.
// This is the CI/pipe path and the most important correctness guarantee.

func TestStatusHelpers_NoTTY_Format(t *testing.T) {
	w := &bytes.Buffer{}
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Success", ui.Success(w, "config: ok"), "✓ config: ok"},
		{"Err", ui.Err(w, "git not found"), "✗ git not found"},
		{"Warn", ui.Warn(w, "cliff changelog: skip"), "! cliff changelog: skip"},
		{"Info", ui.Info(w, "hint: set GH_TOKEN"), "  hint: set GH_TOKEN"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.got)
		})
	}
}

func TestStatusHelpers_MessageAlwaysPreserved(t *testing.T) {
	w := &bytes.Buffer{}
	msg := "git user.name: not configured; run: git config user.name <name>"
	assert.True(t, strings.HasSuffix(ui.Err(w, msg), msg))
	assert.True(t, strings.HasSuffix(ui.Success(w, msg), msg))
	assert.True(t, strings.HasSuffix(ui.Warn(w, msg), msg))
	assert.True(t, strings.HasSuffix(ui.Info(w, msg), strings.TrimLeft(msg, " ")))
}

func TestStatusHelpers_NoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// NO_COLOR must disable styling regardless of writer type.
	w := &bytes.Buffer{}
	assert.Equal(t, "✓ all good", ui.Success(w, "all good"))
	assert.Equal(t, "✗ failed", ui.Err(w, "failed"))
	assert.Equal(t, "! warning", ui.Warn(w, "warning"))
	assert.Equal(t, "  hint: x", ui.Info(w, "hint: x"))
}

func TestHeader_NoTTY(t *testing.T) {
	w := &bytes.Buffer{}
	ui.Header(w, "Runtime")
	got := w.String()
	assert.Contains(t, got, "Runtime")
	assert.True(t, strings.HasPrefix(got, "\n"), "expected leading blank line")
	assert.True(t, strings.HasSuffix(got, "\n\n"), "expected trailing blank line")
}

// CLICOLOR_FORCE=1 forces colorprofile to treat the bytes.Buffer as a color terminal.
func TestStatusHelpers_ColorPath(t *testing.T) {
	t.Setenv("CLICOLOR_FORCE", "1")
	t.Setenv("NO_COLOR", "")

	w := &bytes.Buffer{}
	tests := []struct {
		name  string
		got   string
		plain string
		sym   string
	}{
		{"Success", ui.Success(w, "ok"), "✓ ok", "✓"},
		{"Err", ui.Err(w, "fail"), "✗ fail", "✗"},
		{"Warn", ui.Warn(w, "warn"), "! warn", "!"},
		{"Info", ui.Info(w, "info"), "  info", "info"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, tc.got, tc.sym)
			assert.NotEqual(t, tc.plain, tc.got)
		})
	}
}

func TestHasColor_BufferIsFalse(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("CLICOLOR_FORCE", "")
	assert.False(t, ui.HasColor(&bytes.Buffer{}), "a buffer is not a color terminal")
}

func TestIsTTY_NonFileIsFalse(t *testing.T) {
	assert.False(t, ui.IsTTY(&bytes.Buffer{}), "a non-*os.File writer is never a TTY")
}

func TestIsTTY_RegularFileIsFalse(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "notty")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	assert.False(t, ui.IsTTY(f), "a regular file is not a TTY")
}
