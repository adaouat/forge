# ADR-0001 — Extract the shared `forge` module

**Status:** Accepted
**Date:** 2026-06-01

## Context

`bifrost` (atomic deployment CLI) and `heraut` (release-management CLI) are sibling Go
CLIs under `github.com/adaouat/*`. They were built from the same mental template and have
independently grown near-identical infrastructure:

- Entry point: trivial `cmd/<app>/main.go` → `fang.Execute(ctx, cmd.NewRootCmd())`, version
  injected via `-ldflags "-X main.Version=…"`.
- Stack: `charm.land/{fang,cobra,huh,lipgloss,bubbles}`, `gopkg.in/yaml.v3`.
- Tooling: mise + hk + golangci-lint + goreleaser + cocogitto + typos + yamlfmt + pkl,
  with effectively identical `.config/` trees.
- Process: `docs/specs` + `docs/adr` + `docs/tasks/roadmap.md` (inline `[ ]`/`[x]`),
  `.claude/rules/{workflow,testing,coding,claude}.md`, conventional commits, TDD.

The duplication has already started to drift (diverged copies of the UI status helpers,
different exec-injection seams, slightly skewed dependency pins). `forge` exists to
hold the shared foundation before the drift gets worse.

## Decision

Create `forge` as **both** a runtime Go module (imported by the apps) **and** the
canonical home for shared scaffolding (`.claude/rules`, `.config` tooling, docs/CI
templates).

### Extraction bar

A thing is extracted into forge only if it clears all three:

1. **Identical** — the same code/intent already exists in both apps (or is trivially
   generalizable to both), not merely "similar-looking".
2. **Stable contract** — the public surface is unlikely to churn per-app; a change to it
   is a deliberate, ADR-worthy event.
3. **≥2 real consumers** — bifrost and heraut both use it today, or one uses it and the
   other has a committed near-term need.

This bar is the project's own rule applied to itself: *"three similar lines is better than
a premature abstraction"* and YAGNI. Forge carries **zero domain logic**.

### In scope — runtime packages (Tier 1)

| Package (working name) | Source today | What it holds |
|---|---|---|
| `exec` | heraut `port.Runner` + `adapter/exec`; bifrost `var execCommand` seam | `Runner` interface + real impl (DryRun/Verbose/RunEnv) |
| `exec/exectest` | heraut `testutil` | `MockRunner`, `FakeBin` |
| `ui` | heraut `ui/`, bifrost `tui/` | status (Success/Warn/Err/Info/Header), color+TTY detect, output mode (human/plain/json), version banner, spinner + progress wrappers |
| `exitcode` | bifrost `cmderr`, heraut `exitcode` | `ExitError` + `Resolve`/`Wrap` + `main.go` glue |
| `config` | heraut `config/{loader,path,error}`, bifrost `config/{loader,merge}` | strict YAML loader, path resolution (`--config` → `<APP>_FILE` → `.config/<app>.yml` → `.<app>.yml`), `InitDest`, `ValidationError{Path,Message,Hint}`, merge helpers (`firstNonEmpty`/`concat`/`mergeMaps`) |
| `selfupdate` | heraut `selfupdate` | GitHub Releases API + SHA-256 verify + atomic replace + daily hint, generalized over repo/asset naming |

### In scope — scaffolding (Tier 2)

Canonical, copied into apps with a documented sync source (not a runtime dependency):
`.claude/rules/*`, `.config/{mise,hk,cocogitto,typos,yamlfmt}`, the docs methodology, and
the CI/goreleaser/Dockerfile patterns.

### Explicitly out of scope (Tier 3 — false friends)

- **Config schemas and merge *semantics*.** bifrost is 3-level (`global < env < app`, with
  a `servers` map); heraut is 2-level (`root < env`, with content overrides). The *shape*
  differs — only the primitives are shared, never the schema or the merge tree.
- **Domain logic.** bifrost's hook runner / atomic strategy; heraut's pipeline, generators,
  platforms, versioning. These stay in their apps.

## Consequences

- **One coupling point.** A breaking change to a forge package needs its own ADR and a
  coordinated bump across both apps. Mitigated by keeping packages small and independently
  importable so each app pulls only what it needs.
- **Dependency alignment.** The apps currently pin slightly different versions
  (lipgloss v2.0.1/v2.0.2, bubbles v2.0.0/v2.1.0, cobra 1.9.1/1.10.2). Forge forces a single
  baseline — a one-time reconciliation cost (see roadmap M0).
- **`charm.land` registry convention is preserved** in forge (never
  `github.com/charmbracelet/<module>` as a direct dependency).
- Migration is incremental, one package per roadmap task, with `replace` directives during
  development and local copies deleted only after the shared package is wired and green.
</content>
