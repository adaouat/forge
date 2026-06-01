# adaouat-core — build roadmap

Shared foundation for the `github.com/adaouat/*` CLIs (`bifrost`, `heraut`). The rationale,
extraction bar, and in/out scope live in [ADR-0001](../adr/0001-shared-core-module.md).
This file is the living build plan: each task is `[ ]` until done, then `[x]` with a
one-paragraph note recording actual decisions and deviations.

## Principles (recap)

- **Extraction bar:** identical + stable contract + ≥2 real consumers. Core holds **zero
  domain logic**. When in doubt, leave it in the app.
- **Migration is incremental:** one package per task, behind a `replace` directive, local
  copy deleted only after the shared one is wired and tests are green in both apps.
- **TDD throughout:** failing test first. Every extracted package ships with the tests that
  already cover it in the source app, ported.

## Open decisions (resolve in M0)

- [ ] **Module path** — `github.com/adaouat/core` (clean) vs `github.com/adaouat/adaouat-core`
  (matches the repo dir). Pick one before the first `go.mod`; it is painful to change later.
- [ ] **Dependency baseline** — adopt the newer pin of each skewed dep (lipgloss v2.0.2,
  bubbles v2.1.0, cobra 1.10.2) unless a changelog flags a regression. Record the chosen
  matrix here.
- [ ] **Package naming** — confirm `exec` / `exec/exectest` / `ui` / `exitcode` / `config` /
  `selfupdate`, or adjust. Names are cheap now, expensive after the apps import them.

---

## M0 — Module foundation & scaffold (Tier 2)

- [ ] Initialize the Go module (chosen path, `go 1.26`), `LICENSE.md`, empty `README.md`.
- [ ] Mirror `.config/` tooling from the apps: `mise`, `hk/config.pkl`, `cocogitto`,
      `typos`, `yamlfmt`. Align tool versions (hk 1.46, goreleaser n/a — library).
- [x] Port `.claude/rules/{workflow,testing,coding,claude}.md` as the **canonical** set;
      adapt for a library (no `--output` modes, no deploy specifics). The apps' copies
      become downstream-synced from here. **Done:** merged the union of both apps' rules,
      stripped app-specifics (deploy strategies, containers/`testcontainers`, hexagonal
      layer tables, `--output` modes), and followed heraut's testing model (allows
      `t.TempDir`) over bifrost's container-only one since core needs FS tests for config.
      Added a core-specific "extraction bar / what belongs in core" rule and a "public API
      is a contract" section. Wired into `CLAUDE.md` via `@import`.
- [ ] Create `docs/{specs,adr,tasks}` skeleton; copy ADR-0001 and this roadmap in.
- [ ] CI workflow (`.github/workflows/ci.yml`): build + `go test ./...` + golangci-lint on
      PR. No release workflow yet — core is tagged by hand until v0.1.0.
- [ ] Resolve the **dependency baseline** decision above and pin it in `go.mod`.

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

- [ ] Drop all `replace` directives; tag `adaouat-core` `v0.1.0`.
- [ ] Bump bifrost and heraut to depend on the tagged version; `go mod tidy` both.
- [ ] Per-package contract ADRs in `docs/adr/` (one per Tier-1 package whose interface is
      now load-bearing across two repos).
- [ ] Document the Tier-2 sync workflow (how an app refreshes its `.claude/rules` /
      `.config` from core) in `docs/guides/`.

---

## Explicitly NOT on this roadmap

Per ADR-0001 Tier 3: config **schemas** and **merge semantics**, bifrost's hook runner and
atomic strategy, heraut's pipeline / generators / platforms / versioning. If a future need
makes one of these genuinely shared, it earns its own ADR first — it does not get bolted on
here.
