# ADR-0010 — forge is the CLI framework foundation

**Status:** Accepted
**Date:** 2026-06-05

## Context

forge began as framework-agnostic utilities ([ADR-0001](0001-shared-core-module.md)): the apps
owned the CLI framework (cobra / fang / huh), forge held only lower-level helpers (exec, exitcode,
ui, config, updatecheck), and [ADR-0008](0008-ui-theme-palette.md) deliberately kept the theme out
of forge to avoid a fang/cobra dependency. As the family grows (a 3rd tool, plus a planned
"new project" guide), that boundary now costs more than it saves:

- **Version drift.** Each app pins the framework deps independently — cobra is already split
  (bifrost `1.9.1`, heraut `1.10.2`); fang/huh stay aligned only by luck.
- **Duplication.** The fang theme mapping is copy-pasted per app (ADR-0008's accepted cost).
- forge is, in practice, *the module every tool imports*. Making it the CLI foundation matches
  that reality and gives new tools a batteries-included start.

## Decision

forge is **the family's CLI framework foundation**, not a framework-agnostic utility set. It owns
the shared CLI framework layer; tools are built on it.

- **forge owns fang + huh** (alongside the lipgloss/bubbles it already has) and exposes:
  - `cli.Run(...)` — wraps `fang.Execute`, wiring the version and the family theme, so apps call
    forge instead of fang and **drop their direct fang import** → fang's version is forge's.
  - **the theme** — `ui.ColorScheme` with a forge **default** accent that each tool **overrides**
    (supersedes ADR-0008; the mapping lives once in forge).
  - shared huh helpers/theme as a pattern emerges (apps drop their direct huh import too).
- **cobra stays a direct app dependency.** Apps build their own command trees — theirs to own;
  wrapping cobra's API would be heavy and fight the grain. Its version is **aligned by convention**
  (the Tier-2 `.config`/go.mod baseline; forge drags a floor in via fang). Bump bifrost to heraut's
  `1.10.2` and keep them in lockstep.
- **viper is not adopted** — unused today (config is `forge.Decode` + YAML). Revisit only if a real
  need appears.
- **What does NOT change:** forge still holds **zero domain logic** — no deployment / release /
  versioning logic (ADR-0001's core bar stands). Only the *framework-agnostic* constraint is
  relaxed, and only for the CLI layer.

The MVS reality shapes the mechanism: forge requiring a dep only sets a *floor*, so true
centralization needs the apps to be the **non-direct importer**. That works for fang/huh (reached
via `cli.Run` / forge helpers); it can't for cobra (apps import it to build commands) — hence
"aligned, not wrapped."

## Consequences

- forge's graph gains fang + cobra (+ huh) — heavier, but correct for a foundation library.
- **Drift: fang + huh + theme eliminated; cobra reduced to convention.** A new tool imports forge
  and gets the framework + theme; the planned `docs/guides/new-tool.md` documents that path.
- Supersedes **ADR-0008** (theme now lives in forge) and amends **ADR-0001**'s framework-agnostic
  scope (its zero-domain-logic bar is untouched).
- A pre-stability decision; consumers re-pin forge at the M8 release.
