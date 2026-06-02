# forge — build roadmap

Shared foundation for the `github.com/adaouat/*` CLIs (`bifrost`, `heraut`). The rationale,
extraction bar, and in/out scope live in [ADR-0001](../adr/0001-shared-core-module.md).
This file is the living build plan: each task is `[ ]` until done, then `[x]` with a
one-paragraph note recording actual decisions and deviations.

## Principles (recap)

- **Extraction bar:** identical + stable contract + ≥2 real consumers. Forge holds **zero
  domain logic**. When in doubt, leave it in the app.
- **Migration is incremental:** one package per task, behind a `replace` directive, local
  copy deleted only after the shared one is wired and tests are green in both apps.
- **TDD throughout:** failing test first. Every extracted package ships with the tests that
  already cover it in the source app, ported.

## Open decisions (resolve in M0)

- [x] **Module path** — resolved: `github.com/adaouat/forge`. Chosen over `…/core` and
  `…/adaouat-core`. `adaouat` means "tools" in Arabic, so the org *is* the toolbox and the
  apps (`bifrost`, `heraut`) are finished tools; `forge` names this shared library as *where
  the tools are made*. It's a short, idiomatic leaf with no stutter (`adaouat/forge`, not
  `adaouat/adaouat-core`) and no `core`/stdlib shadowing. Consequence: the GitHub repo must be
  `adaouat/forge` (module path == repo URL for `go get`, barring a vanity import). The
  `adaouat-core` / `Core` name references in CLAUDE.md, ADR-0001, and this file were swept to
  `forge` immediately. Still pending (manual, not yet done): the GitHub repo slug and the local
  dir rename — both land with the M0 `go.mod` init task. The `0001-shared-core-module.md`
  filename is kept as-is (ADR slugs stay stable).
- [x] **Dependency baseline** — resolved: adopt the **newer** pin of each skewed dep, verified
  against both apps' `go.mod`. Chosen matrix: `charm.land/lipgloss/v2` **v2.0.2** (bifrost 2.0.2
  > heraut 2.0.1); `charm.land/bubbles/v2` **v2.1.0** (bifrost 2.1.0 > heraut 2.0.0);
  `github.com/spf13/cobra` **v1.10.2** (heraut 1.10.2 > bifrost 1.9.1). Already aligned:
  `charm.land/huh/v2` v2.0.3, `gopkg.in/yaml.v3` v3.0.1. No changelog flagged a regression.
  The `require` lines land in M1+ as packages import them — at M0 with zero imports `gomod_tidy`
  strips unused pins, so they cannot be pre-pinned. **M3 flag:** heraut depends on
  `github.com/charmbracelet/colorprofile` (not `charm.land`), which `ui` needs — confirm a
  `charm.land` path exists before porting, or document the exception to the registry rule.
- [x] **Package naming** — confirmed as proposed: `exec` / `exec/exectest` / `ui` / `exitcode`
  / `config` / `selfupdate`. No package exists until M1+, so a name stays cheap to change until
  its first import; revisit via ADR if one proves wrong once consumers depend on it.

---

## M0 — Module foundation & scaffold (Tier 2)

- [x] Initialize the Go module (`github.com/adaouat/forge`, `go 1.26.3`), `LICENSE.md`,
      `README.md`. **Done:** `go mod init` pinned the go directive to `1.26.3` (matching heraut
      and the installed toolchain). `LICENSE.md` mirrors the apps' MIT (Brice CHATARD, 2026).
      `README.md` is a short stub (title, one-liner, pointers to ADR-0001 and the roadmap)
      rather than literally empty — an empty file is a worse landing page than three lines.
- [x] Mirror `.config/` tooling from the apps: `mise`, `hk/config.pkl`, `cocogitto`,
      `typos`, `yamlfmt`. Align tool versions (hk 1.46, goreleaser n/a — library). **Done:**
      the tree was copied from heraut, then adapted for a library — dropped the `build`/`run`
      mise tasks (no binary, no `cmd/`), removed `goreleaser` (n/a) and `hadolint` (no
      Dockerfile) from the tool set plus the hadolint linter from `hk/config.pkl`, and
      regenerated `mise.lock` via `mise lock`. Hooks wired via `hk install`.
- [x] Port `.claude/rules/{workflow,testing,coding,claude}.md` as the **canonical** set;
      adapt for a library (no `--output` modes, no deploy specifics). The apps' copies
      become downstream-synced from here. **Done:** merged the union of both apps' rules,
      stripped app-specifics (deploy strategies, containers/`testcontainers`, hexagonal
      layer tables, `--output` modes), and followed heraut's testing model (allows
      `t.TempDir`) over bifrost's container-only one since forge needs FS tests for config.
      Added a forge-specific "extraction bar / what belongs in forge" rule and a "public API
      is a contract" section. Wired into `CLAUDE.md` via `@import`.
- [x] Create `docs/{specs,adr,tasks}` skeleton; copy ADR-0001 and this roadmap in. **Done:**
      `adr/` and `tasks/` already held ADR-0001 and this roadmap; added `docs/specs/` and an
      index `README.md` in each of `docs/`, `adr/`, `specs/`, `tasks/` (mirroring the apps'
      convention, which forge now canonicalizes). The `adr/` index lists ADR-0001 with its
      current status (Accepted). The stray empty `docs/plans/` is left
      untracked (plans live in `.claude/plans/` per `workflow.md`).
- [x] CI workflow (`.github/workflows/ci.yml`): build + `go test ./...` + golangci-lint on
      PR. No release workflow yet — forge is tagged by hand until v0.1.0. **Done:** three jobs
      (lint = golangci-lint + govulncheck, test = `go test ./...` with the apps' 85% coverage
      gate, build = `go build ./...`), adapted from heraut minus the goreleaser/`cmd` steps.
      Actions pinned to commit SHAs reused from heraut. Triggers on push to `main` and PRs; no
      remote exists yet, so it first runs once one is added (code lands in M1).
- [x] Resolve the **dependency baseline** decision above and pin it in `go.mod`. **Done:**
      decision recorded in Open decisions #2 (adopt the newer pins). No `require` added to
      `go.mod` yet — with zero imports, `gomod_tidy` strips unused pins; the baseline applies as
      each dependency is first imported in M1+.

## M1 — `exec` runner + `exectest` (first extraction, lowest risk)

- [ ] `exec.Runner` interface (`Run`, `RunEnv`) + concrete `Runner{DryRun, Verbose, Out}`,
      ported from heraut `internal/adapter/exec`.
- [ ] `exec/exectest`: `MockRunner` (FIFO queued responses, recorded `Calls`) + `FakeBin`,
      ported from heraut `internal/testutil`.
- [ ] Wire into **heraut** behind a `replace` directive: delete `internal/adapter/exec` and
      the runner half of `internal/testutil`, repoint imports. Full suite green.
- [ ] Wire into **bifrost**: replace the `var execCommand = exec.Command` seam in
      `internal/hooks` with the `exec.Runner` interface (the hook runner takes a `Runner`).
      Hook unit tests use `exectest.MockRunner`. Full suite green.

## M2 — `exitcode`

- [ ] `exitcode.ExitError{Code, Message}` + `Resolve(err) int` + `Wrap(code, err)` +
      a `main.go` glue helper, reconciling bifrost `cmderr` and heraut `internal/exitcode`.
- [ ] Migrate heraut `cmd/heraut/main.go` + `internal/exitcode` + `internal/cmd/exit.go`.
- [ ] Migrate bifrost `cmd/bifrost/main.go` + `internal/cmderr`.

## M3 — `ui`

- [ ] Status helpers (`Success`/`Warn`/`Err`/`Info`/`Header`) + `hasColor`/TTY detection via
      `colorprofile`, ported from heraut `internal/ui/status.go`.
- [ ] Output mode (human/plain/json) as an explicit type (generalize bifrost's `tui/mode.go`
      package-global into an injectable value — no shared mutable global in a library).
- [ ] Version banner / header + spinner + progress-bar wrappers.
- [ ] Migrate heraut `internal/ui` and bifrost `internal/tui`; keep app-specific banners
      (ASCII art, colors) as data passed into the shared renderers.

## M4 — `config` primitives

- [ ] Strict YAML loader (`KnownFields(true)`, typed-error formatting) parameterized over a
      target struct.
- [ ] **Path resolution** parameterized over app name: `--config` flag → `<APP>_FILE` env →
      `.config/<app>.yml` → `.<app>.yml`, with the `PathSource` enum and `InitDest`. This is
      the "various file locations" piece — heraut has the reference impl, bifrost gains it.
- [ ] `ValidationError{Path, Message, Hint}` + `ValidationErrors` aggregate.
- [ ] Merge helpers: `firstNonEmpty`, `firstNonZeroInt`, `concat`, `mergeMaps`.
- [ ] Migrate both apps' loaders to the primitives. **Schemas and merge trees stay in the
      apps** (Tier 3) — only the plumbing moves.

## M5 — `selfupdate`

- [ ] Generalize heraut `internal/selfupdate` over repo URL + asset naming pattern (today
      they are compiled-in constants for a single repo).
- [ ] Migrate heraut to the shared package.
- [ ] Wire bifrost's `self-update` command onto it (bifrost gains the feature for free).

## M6 — Finalize & cut v0.1.0

- [ ] Drop all `replace` directives; tag `forge` `v0.1.0`.
- [ ] Bump bifrost and heraut to depend on the tagged version; `go mod tidy` both.
- [ ] Per-package contract ADRs in `docs/adr/` (one per Tier-1 package whose interface is
      now load-bearing across two repos).
- [ ] Document the Tier-2 sync workflow (how an app refreshes its `.claude/rules` /
      `.config` from forge) in `docs/guides/`.

---

## Explicitly NOT on this roadmap

Per ADR-0001 Tier 3: config **schemas** and **merge semantics**, bifrost's hook runner and
atomic strategy, heraut's pipeline / generators / platforms / versioning. If a future need
makes one of these genuinely shared, it earns its own ADR first — it does not get bolted on
here.
