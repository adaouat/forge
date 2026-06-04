@docs/rules/workflow.md
@docs/rules/testing.md
@docs/rules/coding.md
@docs/rules/agent.md

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# CLAUDE.md — forge

Shared foundation for the `github.com/adaouat/*` CLIs — currently `bifrost` (atomic
deployment, `../bifrost`) and `heraut` (release management, `../heraut`). Forge provides the
common CLI runtime packages **and** the canonical scaffolding (`docs/rules`, `.config`
tooling, docs/CI templates) those apps share.

## Status

Shipped. The M0–M6 build plan is complete: forge is a published library (tagged, e.g. `v0.6.x`)
that bifrost and heraut consume off the tag (no `replace`). Packages: `exec` (+ `exectest`),
`exitcode`, `ui`, `config`, `updatecheck`. The public contract is pinned in
[ADR-0007](docs/adr/0007-public-api-surface-and-stability.md); evolution continues per the
extraction bar (ADR-0001) and the change-control rule (breaking change → new ADR + coordinated
bump).

- [ADR-0001](docs/adr/0001-shared-core-module.md) — why forge exists, the extraction bar,
  and what is in/out of scope.
- [docs/tasks/roadmap.md](docs/tasks/roadmap.md) — the build plan (M0–M6) with per-task
  checklist. **Read this before starting any work** and follow the two-step roadmap flow.

## What belongs here (and what doesn't)

Forge carries **zero domain logic**. A thing is extracted only if it clears the bar:
**identical** across both apps + **stable contract** + **≥2 real consumers**. This is YAGNI
applied to the shared layer — *three similar lines beat a premature abstraction*.

- **In:** `exec` (runner + mocks), `ui` (status/modes/spinner/stepper), `exitcode`, `config`
  primitives (loader, path resolution, validation errors), `updatecheck` (version check +
  install-method-aware upgrade hint; no binary self-replacement — see ADR-0005).
- **Out (false friends):** config *schemas* and *merge semantics* (bifrost is 3-level with
  servers; heraut is 2-level with content overrides), bifrost's hook runner, heraut's
  pipeline/generators/platforms/versioning. These stay in their apps.

When tempted to add something, check it against ADR-0001's bar first. If it doesn't clear
the bar, it belongs in the consuming app, not here.

## Conventions

Inherited from the sibling apps (and canonicalized here once M0 lands):
`charm.land` registry for all charmbracelet deps (never `github.com/charmbracelet/<module>`
direct), conventional commits enforced by hk + cocogitto, TDD (failing test first), mise +
hk tooling. These rules now live canonically in `docs/rules/` (ported in M0); the apps still
keep their copies in `.claude/rules/`.
</content>
