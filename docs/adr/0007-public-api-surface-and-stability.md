# ADR-0007 — Public API surface and stability contract

**Status:** Accepted
**Date:** 2026-06-04

## Context

As of M6, bifrost and heraut both import the **published** forge module (the `v0.6.2` tag, no
`replace` directive). Every exported identifier in forge's Tier-1 packages is therefore a
cross-repo contract: a change ripples into two (soon three) repos at the next `go get`. The
coding rules already require that *a breaking change to an exported signature needs an ADR and a
coordinated bump of both consumers* — this ADR pins the load-bearing surface in one place so
that gate is concrete, and points each package at the decision ADR that governs it.

This is **one consolidated ADR**, not one per package: ADRs 0002–0005 already record the
per-package decisions, so separate per-package "contract" ADRs would mostly duplicate them. The
only package without a prior ADR is `config` (M4 shipped without one); its contract is fixed
here. This also matches forge's own YAGNI ethos — three references beat six near-duplicate ADRs.

## Decision

The exported surface below **is** the forge public contract. Each package is load-bearing across
≥2 repos. A **breaking** change (signature or documented behaviour) requires a new ADR
superseding the relevant entry **plus** a coordinated bump of both consumers; **additive**
changes (a new method, a new exported symbol) are minor and need no ADR.

| Package | Public surface (load-bearing) | Governing ADR |
|---|---|---|
| `exec` | `Runner` interface (`Run` / `RunEnv` / `RunDir`), `CmdRunner`, `New` | [0002](0002-exec-runner-working-directory.md) (`RunDir`) |
| `exec/exectest` | `MockRunner` (`Calls`, `QueueResponse`), `Call`, `FakeBin` — the test contract apps assert against | [0002](0002-exec-runner-working-directory.md) |
| `exitcode` | `OK`/`Usage`/`Config`/`Runtime`/`Internal` codes, `ExitError{Code,Message,Err}`, `Resolve`, `Wrap` | [0003](0003-shared-exit-code-vocabulary.md) |
| `ui` | detection (`HasColor`/`IsTTY`), status (`Success`/`Warn`/`Err`/`Info`/`Header`), `Mode`, header renderers (`HelpLong`/`VersionTemplate`), `Spinner` (`Run`/`Step`/`Total`, `Result`, `Skip`) | [0004](0004-ui-spinner-task-runner.md) (`Spinner`) |
| `ui` (theme) | `Palette`/`NewPalette`, `Accent`/`DefaultAccent`, `ColorScheme`, `HuhTheme` | [0008](0008-ui-theme-palette.md), [0010](0010-cli-framework-foundation.md) |
| `config` | `Decode`/`Load` (+ `ErrEmptyConfig`), `Resolver` (`Resolve`/`Label`/`InitDest`, `Source`), `ValidationError`/`ValidationErrors` | — *(fixed here)* |
| `updatecheck` | `Checker.CheckNewer`, `Hinter.Print`, `InstallMethod` + detection | [0005](0005-updates-via-package-managers.md) |
| `cli` | `Run(ctx, cmd, version, accent)` — runs a cobra command through fang with the version + family theme | [0010](0010-cli-framework-foundation.md) |
| `log` | `New(w, level)` — a slog logger rendered via charm.land/log/v2; `LevelFor(verbose)` — the family `--verbose`→level mapping (off→Warn, on→Debug) | [0011](0011-logging-foundation.md) |

The `ui` theme exports and the `cli` package landed in M7/M8 — after this ADR's original date — via
ADRs [0008](0008-ui-theme-palette.md) and [0010](0010-cli-framework-foundation.md). The `log` package
landed in M9 via [ADR-0011](0011-logging-foundation.md) (`New` in v0.10.0, `LevelFor` alongside). All
are additive and folded into the table above so it stays the *complete* surface of record.

Out of scope (per [ADR-0001](0001-shared-core-module.md) Tier 3): app config schemas/merge,
bifrost's hook runner + atomic strategy, heraut's pipeline/generators/platforms/versioning.

## Consequences

- One place to see forge's whole public surface and the decision behind each package.
- Breaking changes are gated (new ADR + coordinated bump). With the `replace`-free dependency,
  a bad bump now surfaces in the apps' `go mod tidy` / CI against the tag, not silently.
- Additive evolution stays cheap — most growth (a new `ui` helper, a new `exitcode` sentinel)
  needs no ADR. The contract is enforced socially (review + this ADR), not by tooling.
- A third consumer inherits the same contract; nothing here is bifrost- or heraut-specific.
