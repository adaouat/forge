package ui_test

import (
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/forge/ui"
)

const (
	testArt    = "██╗  ██╗\n██║  ██║"
	testPhrase = "Every release deserves a héraut."
)

func TestHelpLong(t *testing.T) {
	long := ui.HelpLong(testArt, testPhrase)
	assert.Contains(t, long, testArt, "ASCII art included")
	assert.Contains(t, long, testPhrase, "catch-phrase included")
}

func TestVersionTemplate(t *testing.T) {
	tmpl := ui.VersionTemplate(testArt, testPhrase)
	require.NotEmpty(t, tmpl)
	assert.Contains(t, tmpl, testArt)
	assert.Contains(t, tmpl, testPhrase)
	assert.Contains(t, tmpl, "{{.Name}}")
	assert.Contains(t, tmpl, "{{.Version}}")

	// Must be a valid text/template that renders Name and Version.
	parsed, err := template.New("version").Parse(tmpl)
	require.NoError(t, err, "version template must be valid text/template")

	var buf strings.Builder
	require.NoError(t, parsed.Execute(&buf, struct{ Name, Version string }{"heraut", "1.2.3"}))
	assert.Contains(t, buf.String(), "heraut 1.2.3")
}
