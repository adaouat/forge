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
- **`charmbracelet/log` is the rendering handler.** It implements `slog.Handler` and renders
  CLI-appropriate output (leveled, colored, prefixed) — the natural backend choice given forge
  already standardizes on the `charm.land` family for `ui` (lipgloss, bubbles, huh) and the
  same terminal aesthetic. *(Confirm the `charm.land` vanity path resolves before adding the
  `require`; if it doesn't — same situation as `colorprofile`/`x/term` in the M3 flag — document
  the `github.com/charmbracelet/log` exception in `docs/rules/coding.md`.)*
- **forge exposes setup, not log statements.** Something in the shape of
  `log.New(level Level, accent ui.Accent) *slog.Logger` (exact name/signature TBD at
  implementation time): wires verbosity → level, routes to stderr, and respects
  `ui.IsTTY`/`ui.HasColor` for plain-vs-colored output. Apps call it once, get a logger, and
  write their own domain log statements through it.
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
- Apps that already log ad hoc (`fmt.Println`/`log.Printf`) migrate to the shared
  `*slog.Logger` when they adopt — tracked as the implementation task in the roadmap.
</content>
