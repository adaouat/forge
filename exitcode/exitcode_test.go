package exitcode_test

import (
	"context"
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

func TestExitError_NoMessageOrErr_ReturnsEmptyString(t *testing.T) {
	// A bare &ExitError{Code: N} must not panic dereferencing a nil Err.
	err := &exitcode.ExitError{Code: 1}
	assert.Equal(t, "", err.Error())
}

func TestExitError_MessageTakesPrecedenceOverErr(t *testing.T) {
	// ADR-0013: a summary (Message) displays instead of the full Err, so a
	// caller that already showed the full error elsewhere doesn't repeat it.
	err := &exitcode.ExitError{Code: 1, Message: "summary", Err: errors.New("real")}
	assert.Equal(t, "summary", err.Error())
}

func TestExitError_FallsBackToErrWhenMessageEmpty(t *testing.T) {
	err := &exitcode.ExitError{Code: 1, Err: errors.New("real")}
	assert.Equal(t, "real", err.Error())
}

func TestWrapSummary_Nil_ReturnsNil(t *testing.T) {
	assert.NoError(t, exitcode.WrapSummary(2, nil, "summary"))
}

func TestWrapSummary_DisplaysSummaryNotFullError(t *testing.T) {
	base := errors.New("full detail already shown by a step reporter")
	wrapped := exitcode.WrapSummary(exitcode.Runtime, base, "release failed")

	require.Error(t, wrapped)
	assert.Equal(t, "release failed", wrapped.Error())
}

func TestWrapSummary_UnwrapReachesFullError(t *testing.T) {
	base := errors.New("full detail")
	wrapped := exitcode.WrapSummary(exitcode.Runtime, base, "summary")
	assert.Equal(t, base, errors.Unwrap(wrapped))
}

func TestWrapSummary_ErrorsIsSeesThroughToSentinel(t *testing.T) {
	sentinel := errors.New("sentinel")
	base := fmt.Errorf("step: %w", sentinel)
	wrapped := exitcode.WrapSummary(exitcode.Runtime, base, "summary")
	assert.ErrorIs(t, wrapped, sentinel)
}

func TestWrapSummary_ResolveUsesGivenCode(t *testing.T) {
	wrapped := exitcode.WrapSummary(exitcode.Config, errors.New("bad config"), "summary")
	assert.Equal(t, exitcode.Config, exitcode.Resolve(wrapped))
}

func TestWrapSummary_AlreadyClassified_PreservesInnerCode(t *testing.T) {
	// Mirrors Wrap's first/innermost-classification-wins rule (ADR-0013).
	inner := exitcode.Wrap(4, errors.New("guard"))
	outer := exitcode.WrapSummary(3, inner, "summary")

	assert.Equal(t, 4, exitcode.Resolve(outer))
	assert.Equal(t, "summary", outer.Error())
}

func TestCodes_GenericVocabulary(t *testing.T) {
	// ADR-0003 — shared exit-code vocabulary. Apps extend in 4-69.
	assert.Equal(t, 0, exitcode.OK)
	assert.Equal(t, 1, exitcode.Usage)
	assert.Equal(t, 2, exitcode.Config)
	assert.Equal(t, 3, exitcode.Runtime)
	assert.Equal(t, 70, exitcode.Internal)
}

func TestResolve_UsesNamedDefaults(t *testing.T) {
	assert.Equal(t, exitcode.OK, exitcode.Resolve(nil))
	assert.Equal(t, exitcode.Usage, exitcode.Resolve(errors.New("boom")))
}

func TestCodes_Interrupted(t *testing.T) {
	// 128+SIGINT, the conventional code for a Ctrl-C'd process.
	assert.Equal(t, 130, exitcode.Interrupted)
}

func TestResolve_ContextCanceled_MapsToInterrupted(t *testing.T) {
	assert.Equal(t, exitcode.Interrupted, exitcode.Resolve(context.Canceled))
}

func TestResolve_WrappedContextCanceled_MapsToInterrupted(t *testing.T) {
	wrapped := fmt.Errorf("running command: %w", context.Canceled)
	assert.Equal(t, exitcode.Interrupted, exitcode.Resolve(wrapped))
}

func TestResolve_ExplicitCodeBeatsContextCanceled(t *testing.T) {
	// A command that classifies its own cancellation keeps that classification.
	err := exitcode.Wrap(exitcode.Runtime, context.Canceled)
	assert.Equal(t, exitcode.Runtime, exitcode.Resolve(err))
}
