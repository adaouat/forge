# ADR-0004 — `ui.Spinner` task runner

**Status:** Accepted
**Date:** 2026-06-03

## Context

M3 deliberately **dropped** spinner/progress wrappers, reasoning they were a false friend:
heraut hand-rolls an inline animator over `bubbles` spinner frames inside a `Step`/`Progress`
step-runner, while bifrost uses `huh`'s action-spinner for one operation and a `bubbles`
progress bar for another. No identical code existed, so a shared abstraction looked forced.

Re-examined at the use-case level, the spinner cases *do* share one shape — **"run a named
unit of work, animate a spinner while it runs, then resolve to a `✓`/`✗`/`!` status line"**:

- bifrost `purge` — one task, success/fail.
- heraut `check` (×2) — one task, success/fail/**skip** (advisory warning).
- heraut pipeline reporter — a *sequence* of tasks with `[N/M]` counters and sub-results.

The differences (skip outcome, counters, sub-results, sequence) are parameters of one
abstraction, not different abstractions. A third family CLI will want the same. The earlier
"different libraries" objection was about the *animation engine*, not the contract — and the
engine is an internal detail the shared API can own.

## Decision

Add an action-based task runner to `forge/ui`:

```go
type Result struct { Detail string; Subs []string }   // success payload
func Skip(detail string) error                          // advisory outcome sentinel

type Spinner struct { /* out, mode, counter */ }
func NewSpinner(out io.Writer, mode Mode) *Spinner
func (s *Spinner) Total(n int) *Spinner                 // opt-in [N/total] counter
func (s *Spinner) Run(name string, fn func() (Result, error)) error
```

`Run` animates a spinner titled `name` while `fn` runs, then renders:

- `(Result, nil)` → `✓ [n/N] name — Detail` + indented `Subs`
- `(_, Skip(d))` → `! [n/N] name — d` (returns `nil`; advisory, not a failure)
- `(_, err)` → `✗ [n/N] name — err` (first line; extra lines indented) and returns `err`

**Engine:** forge owns an internal inline animator (the `bubbles` spinner frames + a ticker
goroutine, lifted from heraut and written once), **not** `huh`'s per-task program — the inline
animator is smooth across multi-step sequences and writes status lines to the caller's
`io.Writer` (so they are captured in logs and tests). Animation is gated on
`mode.IsHuman() && IsTTY(out)`; everything else (CI, pipes, JSON/plain, non-TTY) renders the
status line with no animation. Status lines reuse the package's `Success`/`Warn`/`Err`
helpers, so the `✓`/`✗`/`!` vocabulary is consistent family-wide.

This supersedes the "spinner + progress-bar wrappers dropped" note in roadmap M3.

## Consequences

- **One spinner vocabulary** across heraut, bifrost, and the next tool; a single place to
  enhance (elapsed time, quiet mode, …) or fix.
- **heraut** deletes its hand-rolled animator (goroutine/ticker/mutex) and its `Step`/`Progress`
  types, reshaping `check` and the pipeline reporter onto `Spinner.Run`; dry-run maps to
  `Mode.Plain`. This also gives `Mode` (ADR-less, from M3.2) a genuine second consumer.
- **bifrost** routes `purge` through `Spinner`; its deploy step lines unify from `✔` onto the
  shared `✓` so the spun line matches the rest — a deliberate, visible output change.
- **Out of scope:** the determinate **byte progress bar** (bifrost extract) stays in bifrost —
  it is a different primitive with one consumer today, and earns its own extraction only when a
  second consumer appears.
- The action-based API is the contract; the start/stop style heraut used internally does not
  cross the boundary (every real use was action-shaped).
