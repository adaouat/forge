package exitcode_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/forge/exitcode"
)

func TestResolve_Nil_Zero(t *testing.T) {
	assert.Equal(t, 0, exitcode.Resolve(nil))
}

func TestResolve_PlainError_DefaultsToOne(t *testing.T) {
	// An unclassified error defaults to the generic code 1.
	assert.Equal(t, 1, exitcode.Resolve(errors.New("boom")))
}

func TestWrap_Nil_ReturnsNil(t *testing.T) {
	assert.NoError(t, exitcode.Wrap(2, nil))
}

func TestWrap_PreservesMessageAndUnwrap(t *testing.T) {
	base := errors.New("bad config")
	wrapped := exitcode.Wrap(2, base)

	require.Error(t, wrapped)
	assert.Equal(t, "bad config", wrapped.Error())
	assert.Equal(t, base, errors.Unwrap(wrapped))
	assert.Equal(t, 2, exitcode.Resolve(wrapped))
}

func TestResolve_FindsCodeThroughFmtErrorf(t *testing.T) {
	base := exitcode.Wrap(4, errors.New("E001"))
	chained := fmt.Errorf("running pipeline: %w", base)
	assert.Equal(t, 4, exitcode.Resolve(chained))
}

func TestWrap_AlreadyClassified_FirstCodeWins(t *testing.T) {
	// Re-wrapping an already-coded error must not override the original code.
	inner := exitcode.Wrap(4, errors.New("guard"))
	outer := exitcode.Wrap(3, inner)
	assert.Equal(t, 4, exitcode.Resolve(outer))
}

func TestExitError_LiteralWithMessage(t *testing.T) {
	// bifrost-style construction: code + message, no wrapped error.
	err := &exitcode.ExitError{Code: 2, Message: "invalid config"}
	assert.Equal(t, "invalid config", err.Error())
	assert.Equal(t, 2, exitcode.Resolve(err))
	assert.NoError(t, errors.Unwrap(err))
}

func TestExitError_ErrTakesPrecedenceOverMessage(t *testing.T) {
	err := &exitcode.ExitError{Code: 1, Message: "ignored", Err: errors.New("real")}
	assert.Equal(t, "real", err.Error())
}
