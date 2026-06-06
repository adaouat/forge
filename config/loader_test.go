package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/forge/config"
)

type sample struct {
	Name  string `yaml:"name"`
	Count int    `yaml:"count"`
}

func TestDecode_Success(t *testing.T) {
	var s sample
	err := config.Decode(strings.NewReader("name: forge\ncount: 3\n"), &s)
	require.NoError(t, err)
	assert.Equal(t, "forge", s.Name)
	assert.Equal(t, 3, s.Count)
}

func TestDecode_EmptyInputIsClassifiable(t *testing.T) {
	var s sample
	err := config.Decode(strings.NewReader(""), &s)
	require.Error(t, err)
	assert.ErrorIs(t, err, config.ErrEmptyConfig, "empty input maps to ErrEmptyConfig, not a raw EOF")
}

func TestLoad_EmptyFileIsClassifiable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.yml")
	require.NoError(t, os.WriteFile(path, nil, 0o600))

	var s sample
	err := config.Load(path, &s)
	assert.ErrorIs(t, err, config.ErrEmptyConfig)
}

func TestDecode_UnknownFieldRejected(t *testing.T) {
	var s sample
	err := config.Decode(strings.NewReader("name: forge\nbogus: true\n"), &s)
	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "config:"), "errors are prefixed: %q", err.Error())
	assert.Contains(t, err.Error(), "bogus", "the offending field is named")
}

func TestDecode_TypeMismatchFormatted(t *testing.T) {
	// A YAML map into an int field is a yaml.TypeError; its detail is surfaced.
	var s sample
	err := config.Decode(strings.NewReader("count:\n  nested: 1\n"), &s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unmarshal")
}

func TestDecode_MalformedYAML(t *testing.T) {
	var s sample
	err := config.Decode(strings.NewReader("name: : :\n"), &s)
	require.Error(t, err)
}

func TestLoad_Success(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.yml")
	require.NoError(t, os.WriteFile(path, []byte("name: x\ncount: 1\n"), 0o600))

	var s sample
	require.NoError(t, config.Load(path, &s))
	assert.Equal(t, "x", s.Name)
	assert.Equal(t, 1, s.Count)
}

func TestLoad_OpenError(t *testing.T) {
	var s sample
	err := config.Load(filepath.Join(t.TempDir(), "missing.yml"), &s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing.yml")
}
