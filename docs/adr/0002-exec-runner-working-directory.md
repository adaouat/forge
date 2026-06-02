# ADR-0002 — `exec.Runner` gains a working-directory method

**Status:** Accepted
**Date:** 2026-06-02

## Context

[ADR-0001](0001-shared-core-module.md) names the `exec` package's sources as heraut's
`port.Runner` + `adapter/exec` **and** bifrost's `var execCommand` seam — i.e. bifrost is a
planned second real consumer. M1.1 extracted the runner from heraut, whose contract is
`Run(name, args…)` / `RunEnv(env, name, args…)`: heraut always runs git/gh/cog in the
process working directory, so it never needed to set one.

Wiring bifrost (roadmap M1.4) exposed a gap. bifrost's hook runner sets `cmd.Dir` per hook —
each hook runs in the release directory, or in its own `cmd_dir` override — and this is
load-bearing, tested behaviour (`TestRun_CmdDirOverridesWorkingDir`). The M1.1 `Runner`
contract cannot express a working directory, and a `MockRunner` could not record one. Without
closing this gap, bifrost cannot consume `exec.Runner` and the package falls short of
ADR-0001's "≥2 real consumers" bar.

A per-command working directory is a fundamental, non-speculative capability of "executing
external CLI commands", not a bifrost-specific quirk. It clears the extraction bar: a real
consumer needs it today, and the contract addition is stable.

## Decision

Add a third method to the `exec.Runner` interface:

```go
RunDir(dir string, env []string, name string, args ...string) (string, string, error)
```

`RunDir` is the general form; `Run` and `RunEnv` become conveniences for the common
empty-dir cases:

- `Run(name, args…)` ≡ `RunDir("", nil, name, args…)`
- `RunEnv(env, name, args…)` ≡ `RunDir("", env, name, args…)`

An empty `dir` means "the current process working directory" — exactly the prior behaviour,
so heraut's call sites are untouched. `CmdRunner` sets `cmd.Dir` only when `dir != ""`.
`exectest.MockRunner` records `dir` on `Call.Dir` (empty for `Run`/`RunEnv`) so consumers can
assert it.

## Consequences

- **Interface grew from two methods to three.** Adding a method is source-breaking only for
  external *implementers* of `exec.Runner`; forge owns both implementers (`CmdRunner`,
  `MockRunner`), and the consumers (heraut, bifrost) only *use* the interface, so the bump is
  internal to forge. This is the deliberate, ADR-worthy contract change ADR-0001 requires.
- **bifrost becomes a real consumer**, satisfying the ≥2-consumers bar for `exec`.
- `exectest.Call` gains a `Dir` field; existing assertions on `Name`/`Args`/`Env` are
  unaffected (zero value `""` for the dir-less calls).
- The convenience/general split keeps the common call sites terse while giving the one caller
  that needs a directory a first-class way to pass it — no shell `cd` workaround, no global
  seam.
