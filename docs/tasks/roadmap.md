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
      is a contract" section. Wired into `CLAUDE.md` via `@import`. **Relocated** afterward out
      of `.claude/rules/` to `docs/rules/` (renaming `claude.md` → `agent.md`) so `.claude/`
      holds only Claude settings while the conventions live tool-agnostically under `docs/`;
      `CLAUDE.md` imports them from the new path.
- [x] Create `docs/{specs,adr,tasks}` skeleton; copy ADR-0001 and this roadmap in. **Done:**
      `adr/` and `tasks/` already held ADR-0001 and this roadmap; added `docs/specs/` and an
      index `README.md` in each of `docs/`, `adr/`, `specs/`, `tasks/` (mirroring the apps'
      convention, which forge now canonicalizes). The `adr/` index lists ADR-0001 with its
      current status (Accepted). `docs/plans/` is the canonical plans location
      (`plansDirectory` in `.claude/settings.json`, per `workflow.md`); empty for now, so untracked.
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

- [x] `exec.Runner` interface (`Run`, `RunEnv`) + concrete `Runner{DryRun, Verbose, Out}`,
      ported from heraut `internal/adapter/exec`. **Done:** collapsed heraut's `port.Runner`
      interface and `exec.Runner` struct into one package. The interface is `exec.Runner`
      (per `testing.md`'s "`MockRunner` implements `exec.Runner`"); the concrete struct was
      renamed `CmdRunner` (interface and struct can't share the name — user picked `CmdRunner`
      over `OSRunner`/`Cmd`), keeping the `New(dryRun, verbose)`, `Run`, `RunEnv`, `DryRun`,
      `Verbose`, `Out` surface unchanged. All nine of heraut's edge-case rows ported (success,
      dry-run, failure, env propagation, env dry-run, verbose log, verbose output-echo,
      stderr-in-error, no-dangling-colon) plus a `var _ exec.Runner` compliance assertion.
      Tests drive the runner with real `sh -c` commands instead of `exectest.FakeBin` so this
      commit stays green and self-contained — `FakeBin` lands in M1.2 (it can't import this
      package's test the other way around). `testify v1.11.1` pinned (the apps' version) as the
      first real dependency.
- [x] `exec/exectest`: `MockRunner` (FIFO queued responses, recorded `Calls`) + `FakeBin`,
      ported from heraut `internal/testutil`. **Done:** ported `MockRunner`, `Call`,
      `NewMockRunner`, `QueueResponse` and `FakeBin` verbatim (only the doc comment changed:
      "port.Runner" → "exec.Runner"), split across `mockrunner.go` / `fakebin.go` like the
      source, with a `var _ exec.Runner = (*MockRunner)(nil)` assertion. Left behind in
      heraut's `testutil`: `constants.go` (domain binary names like `Cog`/`GitCliff`/
      `Communique`) and the `mock_generator`/`mock_platform` doubles — all domain, none clear
      the extraction bar. The deferred `FakeBin` runner coverage from M1.1 is recovered here:
      `TestFakeBin_installsRunnableScriptOnPath` drives a `CmdRunner` against an installed
      fake binary (exectest's external test imports both `exec` and `exectest`).
- [x] Wire into **heraut** behind a `replace` directive: delete `internal/adapter/exec` and
      the runner half of `internal/testutil`, repoint imports. Full suite green. **Done**
      (heraut commit `493b1d5`): added `replace github.com/adaouat/forge => ../forge` (+ a
      `v0.0.0` require placeholder, dropped at M6 when forge is tagged). Key decision:
      `internal/port.Runner` became a **type alias** `= exec.Runner` rather than being deleted
      — this leaves heraut's ~25 hexagonal `port.Runner` call sites untouched *and* makes
      forge's interface genuinely load-bearing for heraut (what M6's contract ADRs assume),
      vs. mere structural compatibility. The 4 cmd entrypoints just repoint the
      `execadapter` import to `forge/exec` (still call `New`). 22 test files moved off
      `testutil` runner doubles: 18 swapped the import to `exec/exectest`, 4 (pipeline
      reporter/release/changelog tests) kept `testutil` for its domain constants/mocks and
      gained a second import. Deleted `internal/adapter/exec/{runner,runner_test}.go` and
      `internal/testutil/{fakebin,mock_runner}.go`; `constants.go` + `mock_{generator,platform}.go`
      stayed (domain, below the bar). Suite green at 839 tests; `golangci-lint` clean.
- [x] **Extend `exec.Runner` with `RunDir` ([ADR-0002](../adr/0002-exec-runner-working-directory.md)).**
      *(Surfaced mid-M1.4, not in the original plan.)* **Done:** bifrost's hook runner sets
      `cmd.Dir` per hook (tested behaviour), which the M1.1 `Run`/`RunEnv` contract couldn't
      express — so bifrost couldn't consume `exec.Runner` and the package would have fallen
      short of ADR-0001's ≥2-consumers bar. Added `RunDir(dir, env, name, args…)` as the
      general method; `Run`/`RunEnv` now delegate with an empty dir, so heraut is untouched.
      `CmdRunner` sets `cmd.Dir` only when non-empty; `exectest.Call` gained a `Dir` field and
      `MockRunner.RunDir` records it. 5 new tests; suite at 19. ADR-0002 committed first
      (`e9a7fd4`).
- [x] Wire into **bifrost**: replace the `var execCommand = exec.Command` seam in
      `internal/hooks` with the `exec.Runner` interface (the hook runner takes a `Runner`).
      Hook unit tests use `exectest.MockRunner`. Full suite green. **Done** (bifrost commit
      `bd36ca8`): `Run`/`RunInteractive`/`RunWithEvents`/`runOne` now take a `forgeexec.Runner`
      (injected, not a global). The atomic `Deployer` holds one (`forgeexec.New(false,false)`,
      so its constructor signature is unchanged); `activate`/`rollback` build one locally.
      `runOne` calls `RunDir` to preserve per-hook `cmd.Dir`, and recovers the exit code via
      `errors.As(runErr, *exec.ExitError)` since forge wraps with `%w`. The 17 hook tests
      dropped the `os/exec` `TestHelperProcess`/`captureExec` subprocess dance for
      `exectest.MockRunner`; `TestRun_CmdDirOverridesWorkingDir` now asserts `Calls[0].Dir`.
      Behaviour preserved; 153 tests green, `golangci-lint` clean (incl. `-tags integration`).
      forge's error embeds captured stderr (which `runOne` already streams to `out`), so the
      `allow_fail` warning initially echoed stderr twice; fixed in bifrost `29f8dfd` by
      reporting `exit status N` instead, restoring single-stream output and the prior wording.
      **M1 complete:** `exec` + `exec/exectest` now have two real consumers, clearing
      ADR-0001's bar.

## M2 — `exitcode`

- [x] `exitcode.ExitError{Code, Message}` + `Resolve(err) int` + `Wrap(code, err)` +
      a `main.go` glue helper, reconciling bifrost `cmderr` and heraut `internal/exitcode`.
      **Done:** `ExitError` is `{Code, Message, Err}` (user-chosen over `{Code, Err}`+`New`) —
      `Error()` prefers `Err` then `Message`, `Unwrap()` returns `Err`. This serves bifrost's
      `{Code, Message}` struct literals *and* heraut's wrapped-error chains from one type.
      `Wrap` (nil→nil, first-code-wins via `errors.As`) and `Resolve` (nil→0, coded→code,
      else→1) ported from heraut. **No value constants** in forge *(superseded by M2.4 /
      ADR-0003 — the generic core `OK/Usage/Config/Runtime/Internal` was later shared; only
      heraut's `Promotion` stays domain)*. The "main.go glue helper" is `Resolve` itself: both mains collapse to
      `os.Exit(exitcode.Resolve(err))` (forge must never call `os.Exit`, so it returns the code
      only). 8 tests ported + bifrost-shape coverage; suite at 27.
- [x] Migrate heraut `cmd/heraut/main.go` + `internal/exitcode` + `internal/cmd/exit.go`.
      **Done** (heraut commit `cfb2dfe`): `internal/exitcode` is now a thin facade — keeps the
      Spec 01 code values (`Config`/`Runtime`/`Promotion` are domain) and delegates `Wrap`/
      `Resolve` to forge; the local `Error` type was dropped (nothing referenced it directly).
      `main.go` and `internal/cmd/exit.go` were left unchanged: they keep calling
      `exitcode.Resolve`/`Wrap` through the facade, so the ~100 `exitcode.*` call sites and
      `exitcode_test.go` are untouched. No `go.mod` change (forge already wired in M1.3). 839
      tests green.
- [x] Migrate bifrost `cmd/bifrost/main.go` + `internal/cmderr`. **Done** (bifrost commit
      `6a8e2d5`): `internal/cmderr.ExitError` is now a type alias for forge's
      `exitcode.ExitError`, so the ~16 `&cmderr.ExitError{Code, Message}` construction sites,
      the `internal/cmd/errors.go` back-compat alias, and the `*cmderr.ExitError` test
      references all keep working with no edits. `main.go` dropped its hand-rolled
      `errors.As`/`os.Exit(1)` fallback for `os.Exit(exitcode.Resolve(err))`. No `go.mod`
      change (forge wired in M1.4). 153 tests green incl. `-tags integration`. **M2 complete:**
      `exitcode` has two real consumers; both apps' mains share one `Resolve`.
- [x] **Shared exit-code vocabulary ([ADR-0003](../adr/0003-shared-exit-code-vocabulary.md)).**
      *(Added after M2 close; supersedes the "no value constants" decision in M2.1.)* **Done:**
      `0/1/2/3` were already used identically across both apps (bifrost just hadn't named them),
      and a third family CLI is imminent — so forge `exitcode` now defines the generic vocabulary
      `OK/Usage/Config/Runtime/Internal`, and `Resolve` returns `OK`/`Usage` by name. Domain codes
      stay in the apps (heraut's `Promotion=4`); forge owns `0–3` + `70`, apps extend in `4–69`.
      2 tests added (suite at 29). ADR-0003 committed first (`ffc3704`).
- [x] Adopt in **heraut**: `internal/exitcode` re-exports forge's generic codes, keeps
      `Promotion=4` as its domain code. Suite green. **Done** (heraut commit `d70b4b3`): the
      const block now reads `OK = forgeexit.OK` … `Internal = forgeexit.Internal`, with
      `Promotion = 4` kept as heraut's domain code; call sites and `exitcode_test.go`
      (incl. `TestCodes_MatchSpec`, which now also guards that forge's values match Spec 01)
      unchanged. 839 tests green.
- [x] Adopt in **bifrost**: `internal/cmderr` re-exports forge's generic codes; replace the
      raw `1/2/3` literals at the construction sites with named constants. Suite green.
      **Done** (bifrost commit `b04ab9c`): `cmderr` re-exports `Usage/Config/Runtime` from forge
      (and `cmd/errors.go` re-exports them again for package-`cmd`-local use by `deploy.go`,
      which uses the bare `ExitError` alias). All ~16 construction sites now use named codes.
      153 tests green incl. `-tags integration`. **Shared exit-code vocabulary complete:**
      three real consumers of the generic set once tool-3 lands; two today.

## M3 — `ui`

**Scope (revised after exploration — user picked "detection + header + status + mode").** The
truly-identical shared surface is narrower than first planned, so two adjustments: **the
progress bar is dropped** (heraut has none; bifrost's byte bar is a determinate, single-consumer
primitive — revisit when a second consumer appears). The **spinner was initially dropped too**,
but on review all spinner uses share one shape (run a named task → `✓`/`✗`/`!`) and were
extracted as `ui.Spinner` — see **M3.6** / [ADR-0004](../adr/0004-ui-spinner-task-runner.md).
**Output mode** is bifrost-only today but is extracted
as a de-globalized value type (canonical for the family + the incoming tool-3); **status
helpers** are heraut-led, bifrost adopts them only where its deploy formats fit. **Dependency
note:** `colorprofile`/`x/term` have *no* usable `charm.land` module path (their `go.mod`
declares `github.com/charmbracelet/*`; the vanity path resolves a version but can't be
`require`d), so forge imports the `github.com` paths — a documented exception to the registry
rule (`docs/rules/coding.md`). `lipgloss/v2` uses `charm.land` as normal.

- [x] Color/TTY detection (`HasColor`, `IsTTY`) + status helpers (`Success`/`Warn`/`Err`/
      `Info`/`Header`), ported from heraut `internal/ui/status.go`. **Done:** ported the five
      status helpers verbatim (`hasColor` → exported `HasColor`); `IsTTY` generalizes heraut's
      `isTerminal` (the `*os.File` check from `step.go`) to any `io.Writer`. Deps pinned:
      `charm.land/lipgloss/v2` v2.0.2 (baseline), `github.com/charmbracelet/colorprofile` v0.4.2
      + `github.com/charmbracelet/x/term` v0.2.2 (the registry exception). 16 tests (heraut's
      status rows + `HasColor`/`IsTTY` coverage); forge suite at 45.
- [x] Output mode (human/plain/json) as an injectable `Mode` value type (de-globalize bifrost's
      `tui/mode.go` package-global — no shared mutable global in a library). **Done:** `Mode int`
      with `Human`/`Plain`/`JSON` (zero value `Human`, matching bifrost's default), `ParseMode`
      (unknown → `Human`), `IsHuman`, and `String` (round-trips `ParseMode`). A plain value, so
      bifrost holds/injects one instead of the `atomic.Int32` package-global. 10 tests; ui at 26.
- [x] Version banner / header renderers (`HelpLong(art, phrase)`, `VersionTemplate(art,
      phrase)`) — app-specific ASCII art + catch-phrase passed in as data. *(Spinner +
      progress-bar wrappers dropped — see scope note.)* **Done:** both apps' `header.go` had
      byte-identical `HelpLong`/`VersionTemplate` differing only in the `asciiArt`/`CatchPhrase`
      constants, so the renderers move to forge parameterized over those two strings; apps keep
      the constants and pass them. 2 tests (incl. rendering `{{.Name}} {{.Version}}`); ui at 28.
- [x] Migrate heraut `internal/ui`: route status/header/detection through forge; keep the
      step-runner, progress, and `asciiArt`/`CatchPhrase` data in heraut. **Done** (heraut
      `20db1c2`): `status.go` is thin wrappers over forge; `header.go` keeps the art/phrase and
      calls forge's renderers; `step.go` drops its local `isTerminal` for `forge.IsTTY`. The
      ~20 external `ui.*` call sites in `cmd` are untouched. 839 tests green.
- [x] Migrate bifrost `internal/tui`: route header + detection + de-globalized `Mode` through
      forge; keep spinner, progress bar, deploy UI, and styles in bifrost. **Done** (bifrost
      `961035e`): the `atomic.Int32` output-mode global is gone — the deploy/progress renderers
      take a `forgeui.Mode`, the `Deployer` holds one (replacing `jsonMode`, now a method),
      and `cmd` builds it via `forgeui.ParseMode(--output)` so `root` sets no global. `header.go`
      wraps forge; `IsTTY` delegates to forge. Spinner/progress/styles/JSON-emitter stay. 152
      tests green incl. `-tags integration`. **M3 complete:** `ui` has two real consumers;
      detection + header are genuinely shared, status/mode are canonical for tool-3.

### M3.6 — `ui.Spinner` (spinner extraction, [ADR-0004](../adr/0004-ui-spinner-task-runner.md))

*(Re-opens the spinner the scope note initially dropped. On review all spinner uses are one
shape — run a named task, animate, resolve to `✓`/`✗`/`!` — so it's extracted, enhanced, and
shared across both apps + tool-3.)*

- [x] forge `ui.Spinner`: action-based `Run(name, fn) (✓/✗/!)`, `Result{Detail,Subs}`, `Skip`
      sentinel, opt-in `Total(n)` `[N/M]` counter; inline `bubbles`-frames animator gated on
      `Mode.IsHuman() && IsTTY(out)`, status lines via the existing helpers. ADR-0004 first.
      **Done:** the animator (ticker goroutine + mutex/`\r`-clear) is lifted from heraut's
      `step.go` and written once; `render` reproduces heraut's exact status lines (`✓ name`,
      `✓ name — detail`, multi-line `✗`, `!`, indented subs, `[N/M]`). Tests use a buffer
      (non-TTY → no animation, deterministic). `charm.land/bubbles/v2` v2.1.0 pinned (baseline).
      9 tests; forge suite at 66. ADR-0004 committed first (`d5493c2`).
- [x] Migrate heraut: delete the hand-rolled animator + `Step`/`Progress`; reshape `check.go`
      (×2) and the pipeline reporter onto `Spinner.Run`. **Done** (heraut `d46b7fd`): `step.go`
      + its tests deleted; `check.go`'s start/stop sites became `Spinner.Run` closures
      (`Result`/`err`/`Skip`); `app.spinnerReporter` adapts `ui.StepFn` (kept) to `Spinner` with
      a `Total` counter. Correction to the plan: `Mode.Plain` for dry-run was **not** needed —
      `StartPlainStep` was dead code, so heraut always spun-on-TTY; preserved by passing
      `Mode.Human`. The unreachable nil-fn guard was dropped (pipeline always passes a real fn).
      822 tests green.
- [x] Migrate bifrost: `purge` → `Spinner.Run`; unify deploy step lines (`✔` → `✓`). The
      byte progress bar stays in bifrost (single-consumer, determinate — out of scope).
      **Done** (bifrost `4a3b68f`): `purge` spins + resolves via `Spinner.Run` in the non-JSON
      branch (JSON still emits events); `tui.RunWithSpinner` removed; `PrintStep` renders via
      forge's `Success` (`✔`→`✓`, flush-left), dropping its `Mode` param. `DeployHeader`/
      `PrintSummary`/`PrintDetail`/progress bar stay. 152 tests green incl. `-tags integration`.
      **M3.6 complete:** one spinner vocabulary across both apps + tool-3; `Mode` gained a
      second consumer (heraut passes `Human`).

### M3.7 — numbered deploy steps (bifrost adopts the `[N/total]` stepper)

*(The `[N/total]` step-runner already shipped as `Spinner.Total` in M3.6; this wires bifrost's
deploy onto it. Adding `Spinner.Step` because bifrost's steps are already-completed work, not
fn-to-run.)*

- [x] forge `Spinner.Step(name, detail)`: render-only numbered `✓ [N/M] name — detail` for a
      completed step, sharing `Run`'s counter via a `nextLabel` helper. **Done:** success-only
      (failures surface before the line); 4 tests, ui at 41. ADR-0004 amended with the
      Run-vs-Step distinction.
- [x] bifrost: pre-compute the deploy step total (7 base + non-empty hook groups), thread one
      `Spinner.Total` through `Deploy`, render each step via `Step` (purge still via `Run` on the
      same spinner). Remove the now-unused `tui.PrintStep`. **Done** (bifrost `3d12e7b`):
      `deployStepTotal` counts the always-present 7 + configured hook groups; human-mode lines
      now read `✓ [N/M] …`, JSON mode untouched. `tui.PrintStep` + its tests removed; 150 tests
      green incl. `-tags integration`. **M3.7 complete:** both apps now share the `[N/total]`
      stepper (`Spinner.Total` + `Run`/`Step`).

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
      `.config` from forge's canonical `docs/rules` / `.config`) in `docs/guides/`.

---

## Explicitly NOT on this roadmap

Per ADR-0001 Tier 3: config **schemas** and **merge semantics**, bifrost's hook runner and
atomic strategy, heraut's pipeline / generators / platforms / versioning. If a future need
makes one of these genuinely shared, it earns its own ADR first — it does not get bolted on
here.
