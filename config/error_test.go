package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adaouat/forge/config"
)

func TestValidationError_NoHint(t *testing.T) {
	e := config.ValidationError{Path: "versioning.strategy", Message: "unknown strategy"}
	assert.Equal(t, "versioning.strategy: unknown strategy", e.Error())
}

func TestValidationError_WithHint(t *testing.T) {
	e := config.ValidationError{Path: "x", Message: "bad", Hint: "set y"}
	assert.Equal(t, "x: bad\n  hint: set y", e.Error())
}

func TestValidationErrors_JoinsWithNewlines(t *testing.T) {
	ve := config.ValidationErrors{
		{Path: "a", Message: "m1"},
		{Path: "b", Message: "m2", Hint: "h"},
	}
	assert.Equal(t, "a: m1\nb: m2\n  hint: h", ve.Error())
}

func TestValidationErrors_IsError(t *testing.T) {
	var err error = config.ValidationErrors{{Path: "a", Message: "m"}}
	assert.Error(t, err)
}
