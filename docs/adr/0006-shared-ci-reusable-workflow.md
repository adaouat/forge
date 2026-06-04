# ADR-0006 — Shared lint/test CI via a reusable workflow in forge

**Status:** Accepted
**Date:** 2026-06-04

## Context

forge canonicalizes the family's shared scaffolding ([ADR-0001](0001-shared-core-module.md)
Tier-2: `docs/rules`, `.config` tooling, CI templates). The three repos (forge, bifrost,
heraut) each carry a near-identical CI `ci.yml` whose **lint** job (`golangci-lint` +
`govulncheck`) and **test** job (`go test ./...` + a coverage gate) are byte-identical apart
from the coverage threshold (bifrost 20, forge/heraut 85). They drift independently — action
SHAs, tool setup, and lint flags must be bumped in three places. (CI runs no integration-tagged
tests anywhere: those sit behind a build tag and are excluded by plain `go test ./...`.)

Two ways to de-duplicate:

- **Synced template** — copy forge's `ci.yml` into each repo, like `.config`/`docs-rules`.
  Copy-drift returns the moment one repo edits its copy.
- **Reusable `workflow_call` workflow** — referenced natively by each repo. One source of
  truth, no drift; the trade-off is a *live* dependency (a repo's CI resolves a workflow from
  forge at run time).

The **build** job genuinely differs (forge `go build ./...`, no goreleaser; bifrost/heraut
`go build ./cmd/<app>/` + `goreleaser check` + `fetch-depth: 0`) and is **out of scope** — as
are the release workflows (heraut self-releases + a Docker image; bifrost self-releases only).

## Decision

Host a reusable **lint + test** workflow at `forge/.github/workflows/go-ci.yml`
(`on: workflow_call`), consumed by all three repos — **forge included** (it calls its own via
`uses: ./.github/workflows/go-ci.yml`), giving three real consumers.

- **Scope: lint + test only.** Lint = `golangci-lint run ./...` + `govulncheck ./...`. Test =
  `go test ./... -coverprofile` + a coverage-threshold gate. Build and release stay per-app.
- **One input:** `coverage-threshold` (number), passed to the gate via an `env:` var (never
  interpolated into the shell). It is **required with no default** — the threshold is
  per-project policy (bifrost 20, heraut/forge 85; not a shared convention), so forge holds no
  default that could silently govern a caller or move its gate when forge changes. A caller that
  omits it fails immediately with a clear "missing required input" error.
- **No "which repo" input.** A called reusable workflow's `actions/checkout` resolves the
  *caller's* repository and SHA, so the workflow lints/tests whichever repo invokes it, with
  that repo's mise-provisioned tools.
- **Caller refs pin a forge commit SHA** (forge is untagged until M6), consistent with the
  SHA-pinning rule; the same-repo (forge) call uses the local `./…` form.
- **forge must be a public GitHub repo** for cross-repo calls to resolve (it will be); no
  private-repo org-access setting is needed.

## Consequences

- A single place to bump action SHAs, lint flags, the Go toolchain policy, and the coverage
  gate for the whole family; the incoming third tool inherits CI by adding one `uses:` line.
- **New coupling:** each app's CI now depends on forge's repo + the pinned workflow SHA being
  available — unlike synced scaffolding, which the app owns outright. A bad ref or a broken
  forge workflow breaks the callers' CI. Mitigated by SHA-pinning (callers upgrade
  deliberately) and the narrow lint/test scope.
- **App wiring waits on forge being published.** Until `adaouat/forge` exists publicly, the
  bifrost/heraut callers cannot resolve the workflow, so they are wired in the same change that
  publishes forge (M6-adjacent). The forge-side workflow + forge's own dogfooding call land now.
- Build/release stay per-app, so this does not entangle forge with goreleaser, Docker, or
  release ownership.
