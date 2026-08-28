# ADR-0013 — `exitcode.ExitError` summary/full-error split

**Status:** Accepted
**Date:** 2026-08-28

## Context

heraut (built on `forge/cli` + `forge/ui` + `forge/exitcode`) hit a UX bug that traces back to
this package: when a `forge/ui.Spinner`-reported step fails, the error is displayed twice — once
inline by `Spinner.Run`'s own render (`label + " — " + err.Error()`), and a second time verbatim
in fang's boxed "ERROR" panel, because `fang.Execute` unconditionally calls its `errHandler` on
whatever a command's `RunE` returns, and `cli.Run` (ADR-0010) exposes no per-error suppression or
customization hook.

`exitcode.ExitError` already carries a `Message` field alongside `Err` — bifrost's literal
construction sites set `Message` with no `Err`, `Wrap` sets `Err` with no `Message` — but
`Error()` always prefers `Err` once it's set:

```go
func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}
```

So there is no way to construct an error that displays a short summary at the top level (via
`Error()`) while keeping the full underlying error reachable via `Unwrap` for `errors.Is`/
`errors.As` classification downstream (heraut's own `RunE`-boundary wrapping checks
`errors.Is(err, someSentinel)` before assigning a code). This gap is generic to any tool
combining `forge/ui.Spinner` with `forge/cli.Run` — not heraut-specific — so it belongs here
rather than as a per-tool wrapper reinvented in every consuming CLI.

Grepped every `ExitError{...}` literal and `Wrap(...)` call site in both consumers (heraut,
bifrost): none ever sets both `Message` and `Err` on the same value today (bifrost: `Message`
only; heraut: `Wrap`, which sets `Err` only). Flipping the precedence therefore changes zero
observed behaviour right now — but the precedence is documented and directly tested
(`TestExitError_ErrTakesPrecedenceOverMessage`), so per ADR-0007's rule ("a breaking change —
signature or documented behaviour — requires a new ADR"), this is treated as a deliberate,
ADR-gated change rather than a silent edit.

## Decision

1. Flip `Error()`'s precedence: prefer `Message` when non-empty, else `Err.Error()`, else `""`.
2. Add `WrapSummary(code int, err error, summary string) error`: like `Wrap`, but also sets
   `Message`. Mirrors `Wrap`'s "first/innermost classification wins" rule — if `err` already
   carries a code anywhere in its chain, that code is preserved (the summary still attaches on
   top of it).
3. `TestExitError_ErrTakesPrecedenceOverMessage` is renamed and inverted to
   `TestExitError_MessageTakesPrecedenceOverErr`, asserting the new precedence — the deliberate-
   behaviour-change exception `docs/rules/testing.md` requires an ADR for.

Usage — the motivating case is a command's `RunE`, after a step reporter has already shown the
full error, wrapping it with a short summary before returning:

```go
if err := pipeline.Run(ctx); err != nil {
	return exitcode.WrapSummary(exitcode.Runtime, err, "release failed")
}
```

fang's error panel then renders "release failed" instead of repeating the full error; anything
downstream classifying the error (`errors.Is`/`errors.As`) still sees through to the original.

## Consequences

- **Not breaking in practice.** Verified (grep, both consumers) that no existing bifrost or
  heraut call site sets both `Message` and `Err`, so every current caller is unaffected. Per
  ADR-0007's letter, a documented-behaviour change still technically calls for "a coordinated
  bump of both consumers"; this ADR treats that as satisfied by the empirical zero-impact check
  rather than forcing an immediate heraut/bifrost bump — heraut's own adoption of
  `WrapSummary` at its `RunE` boundaries is a separate, later change.
- `Wrap` itself is untouched — every existing caller keeps its exact behaviour and resolved code.
- This does not otherwise address the `cli.Run`/fang double-display root cause (no `Spinner` or
  `fang` changes) — it only gives callers the tool to avoid it. Wiring it into heraut's `RunE`
  is out of scope here, tracked as a heraut-side follow-up.
- ADR-0007's `exitcode` row gains `WrapSummary` and this ADR as a governing reference.
