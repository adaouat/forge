# ADR-0011 — Logging foundation: `slog` API, `charmbracelet/log` handler

**Status:** Accepted
**Date:** 2026-06-08

## Context

forge owns the family's CLI framework foundation ([ADR-0010](0010-cli-framework-foundation.md)):
`cli.Run` (fang), the theme (`ui.ColorScheme` / `ui.HuhTheme`), and TTY/color detection
(`ui.IsTTY`, `ui.HasColor`). Logging setup is the same shape of problem ADR-0010 already
solved for the framework layer — each app would otherwise wire its own verbosity flags,
level mapping, and output handler, drifting the way fang/huh/the theme did before M8.

Two concerns get conflated when "logging" comes up, and only one of them clears the
[ADR-0001](0001-shared-core-module.md) bar:

- **Logging *content*** — what an app logs, with what attributes (`samber/oops`-style
  structured error context, domain fields, routing to Loki/Sentry/Datadog). This is
  inherently app-specific: it lives in the domain layer ADR-0001 explicitly keeps out of
  forge. Not identical across consumers, not a stable shared contract.
- **Logging *setup*** — mapping `--verbose`/`--quiet` to levels, choosing a handler, routing
  output to stdout/stderr with TTY-aware formatting. This is mechanical plumbing, identical
  in shape across `bifrost`/`heraut`/the next tool, and a sibling to what `ui` and `cli.Run`
  already centralize.

Only the second clears the bar.

## Decision

forge gains a thin logging-setup package built on two layers that play the same
interface/implementation roles as `cli.Run`/fang:

- **`log/slog` (stdlib) is the API.** All forge and app code logs through `*slog.Logger` —
  zero dependency cost, and it is the common interface every backend (including
  `charmbracelet/log`, `zerolog`, the `samber/slog-*` sinks) already targets. Forge does not
  invent its own logging facade.
- **`charm.land/log/v2` is the rendering handler.** It implements `slog.Handler` and renders
  CLI-appropriate output (leveled, colored, prefixed) — the natural backend choice given forge
  already standardizes on the `charm.land` family for `ui` (lipgloss, bubbles, huh) and the
  same terminal aesthetic. *(Verified: the v2 line is properly republished under `charm.land`
  per the project's [UPGRADE_GUIDE_V2](https://github.com/charmbracelet/log/blob/main/UPGRADE_GUIDE_V2.md)
  — `charm.land/log/v2` resolves and `go get`s cleanly, no `colorprofile`/`x/term`-style
  exception needed; the unversioned `charm.land/log` is a stale v0/v1 proxy entry that still
  declares `github.com/charmbracelet/log` and fails on `go get`.)*
- **forge exposes setup, not log statements.** **As built (M9):**
  `log.New(w io.Writer, level slog.Level) *slog.Logger` (the constructor) and
  `log.LevelFor(verbose bool) slog.Level` (the family `--verbose`→level mapping). Apps build a
  logger once per command — `log.New(os.Stderr, log.LevelFor(verbose))` — and inject it into
  their domain services; forge owns the mapping and the rendering, the app owns the flag and
  the call sites. No `ui.IsTTY`/`ui.HasColor` wiring is needed — `charm.land/log/v2` runs
  `colorprofile.Detect` internally.
- **Out of scope (stays in the apps, or doesn't happen at all):**
  - Routing to external sinks (Loki, Sentry, Datadog, Kafka, …) — a deployed-service concern,
    not a CLI runtime concern. If a tool needs it, it wires `samber/slog-*` sinks itself on top
    of the `*slog.Logger` forge hands back.
  - Structured error-context enrichment (`samber/oops` or similar) — domain logic; an app's
    call, not forge's. forge's own error handling stays the existing `fmt.Errorf("...: %w", err)`
    + sentinel/typed-error convention (`docs/rules/coding.md`); adopting `oops` would be a
    breaking change to that contract and needs its own ADR if ever proposed.

## Consequences

- forge's graph gains `charmbracelet/log` (pending the `charm.land` path check above) —
  consistent with the ADR-0010 framing: a foundation library is expected to carry the family's
  shared runtime deps so the apps don't each pin them adrift.
- A new tool gets logging "for free" the same way it gets `cli.Run` + the theme — one call,
  consistent verbosity flags and output across the family.
- Does not relax forge's zero-domain-logic bar (ADR-0001): only the mechanical setup moves
  here; what gets logged, and where it ultimately lands in production, stays the apps' call.
- **Finding (M9.3):** neither bifrost nor heraut had *any* ad hoc logging to migrate — their
  `--verbose` drives `forge/exec` command-echoing, and their user output goes through `ui` /
  stdout, not a logger. So the first real use is **new operator-debugging logging** (leveled
  diagnostics on stderr, gated by `--verbose`), not a migration. heraut is the first adopter;
  bifrost is deferred until it grows a need. The mapping (`LevelFor`) is a deliberate
  family-wide contract, so it lives in forge even though heraut is its first consumer.
</content>
