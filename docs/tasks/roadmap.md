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

**Scope (revised after exploration — user picked "loader + path + ValidationError").** The
loader and path resolver are genuinely shared (both apps, identical patterns). `ValidationError`
is heraut-led but extracted as canonical structured errors (bifrost adopts it over its
`[]string`). **Merge helpers (`firstNonEmpty`/`firstNonZeroInt`/`concat`/`mergeMaps`) are
dropped** — they're bifrost-only; heraut's merge is domain (`MergeContentDriver`) and its
`mergeMaps` is a different deep-merge, so extracting them would be a 1-consumer abstraction.
They stay in bifrost. Schemas, defaults/normalize, and merge trees stay in the apps (Tier 3).

- [x] Strict YAML loader (`KnownFields(true)`, typed-error formatting) parameterized over a
      target struct. **Done:** `Decode(r, target any)` (strict decode; `yaml.TypeError`
      flattened to a joined message) + `Load(path, target any)` (open + Decode).
      Idiomatic non-generic form (`target any`) — apps keep their `*Config` wrappers + defaults.
      forge owns the error wording so both apps emit identical text: `Decode` prefixes `config:`,
      `Load` prefixes `open config %q:`. 7 tests; `yaml.v3` v3.0.1 pinned direct.
- [x] **Path resolution** parameterized over app name: `--config` flag → `<APP>_FILE` env →
      `.config/<app>.yml` → `.<app>.yml`, with a `Source` enum, `Label`, and `InitDest`.
      `Resolver{App}`; heraut has the reference impl, bifrost already matches it. **Done:**
      `Resolver{App}.Resolve(explicit) (path, Source)` with `FromFlag/FromEnv/FromXDG/FromDefault`;
      `Label(src)` rebuilds heraut's "(from HERAUT_FILE)" display; `InitDest()` is the
      `.config/`-check. No global `statFile` seam (bifrost had one) — tests use
      `t.Chdir(t.TempDir())` + `t.Setenv`. 9 tests; config at 15.
- [x] `ValidationError{Path, Message, Hint}` + `ValidationErrors` aggregate (ported from
      heraut; bifrost adopts over `[]string`). **Done:** ported verbatim from heraut's
      `error.go` (`Path: Message` + optional `\n  hint:`; aggregate joins with `\n`). 4 tests;
      config at 19.
- [x] Migrate heraut `internal/config` loader/path/error to forge facades; keep schema,
      validator, normalize, `MergeContentDriver`. **Done** (heraut `c4dfca5`): `LoadFromReader`
      uses `forge.Decode` (heraut keeps its `config:` prefix + `normalize`); `ResolvePath*`/
      `InitDest` wrap `Resolver{App:"heraut"}` — `PathSource` + constants stay (values match
      forge's `Label`, so `check.go`/`path_test` untouched); `ValidationError`/`ValidationErrors`
      become type aliases. 822 tests green.
- [x] Migrate bifrost `internal/config` loader + `cmdutil` path resolution to forge; adopt
      `ValidationError`. Keep schema, 3-level merge, defaults, merge helpers. **Done** (bifrost
      `38feebb` + `58bc083`): `Parse` uses `forge.Decode`; `ValidateServerRefs`/`Validate` return
      `forge.ValidationErrors` (7 call sites build the message via `errs.Error()`); `cmdutil`
      path resolution delegates to `Resolver{App:"bifrost"}` (the `StatFile` seam stays for
      config-init's existence check; path tests moved to `t.Chdir`). 146 tests green incl.
      `-tags integration`. **M4 complete:** loader + resolver + `ValidationError` shared; both
      apps' schemas/merge/defaults stay domain.

## M5 — `updatecheck` (package-manager-delegated upgrades, [ADR-0005](../adr/0005-updates-via-package-managers.md))

**Scope (revised after exploration — user picked the package-manager route).** No binary
self-replacement: heraut's hand-rolled updater (GitHub fetch + SHA-256 + atomic `os.Rename`)
is bug-prone and the Go self-update lib landscape is thin (creativeprojects/go-selfupdate = 51
modules incl. go-github + GitLab + Gitea SDKs; minio/selfupdate carries maintainer-trust risk;
inconshreveable/go-update is zero-dep but 2016-frozen). The project already standardises on
**mise** + **goreleaser**, which install/upgrade GitHub-release binaries with no custom code.
So forge owns only the safe half (check + install-detection + hint); the upgrade itself is
delegated to the package manager. Renamed `selfupdate` → `updatecheck` (it performs no update).

- [x] forge `updatecheck`: GitHub `releases/latest` check + install-method detection + 24h hint.
      **Done:** `Checker{Repo,BaseURL,Client}.CheckNewer(ctx, current)` (GET releases/latest,
      httptest-tested, never real GitHub); `isNewer`/`compareVersions` ported from heraut
      (component-wise numeric — SemVer *and* CalVer, the `v1.10.0 > v1.9.0` edge case);
      `DetectInstall` + `InstallMethod.UpgradeCommand(bin, module)` (Homebrew `/Cellar/`, mise
      `/mise/installs/`, Scoop `\scoop\`, go-install `$GOBIN`/`$GOPATH/bin`; `""` → caller's
      generic fallback), with a path-only `detect(path, goBin)` core that's fully testable;
      `Hinter.Print` (24h cache + `Now func()` clock seam, all errors swallowed). **Stdlib-only**
      (`net/http`+`encoding/json`+`os`/`filepath`) — no new dependency. 29 tests; forge at 118.
- [x] Migrate heraut: delete `internal/selfupdate` (the binary replacer); the daily hint comes
      from forge's `updatecheck`; the `self-update` command is **dropped** (user's call — the
      check is enough). **Done** (heraut `b92ff7a`): removed the whole `internal/selfupdate`
      package + the `self-update` command (net −1050 lines); `PersistentPostRunE` now runs
      `updatecheck.Hinter{Repo,Bin,Module,Current,CacheFile}` (cached 24h under `UserCacheDir`,
      500ms timeout, errors swallowed), still gated on dev builds / `HERAUT_CHECK_UPDATE=false`.
      `NewRootCmd` lost its `selfupdate.Option` param. 785 tests green. *(`--version` can't carry
      the check — fang/cobra short-circuit the flag before any hook, and heraut overloads
      `--version` on subcommands; the cached daily hint on normal commands is the mechanism.)*
- [x] Wire bifrost: gains the update-check hint + install-method-aware upgrade message for free.
      **Done** (bifrost `refactor(cmd): wire forge updatecheck hint and version injection`):
      bifrost had **no version injection** at all (`fang.Execute` was called without a version,
      and `.goreleaser.yml` ldflags lacked `-X main.Version`), so this added `var Version = "dev"`
      in `cmd/bifrost/main.go`, threaded it through `cmd.NewRootCmd(version)` + `fang.WithVersion`,
      and runs `updatecheck.Hinter{Repo:"adaouat/bifrost", Bin:"bifrost", Module, Current, CacheFile}`
      in a new `PersistentPostRunE`. Gated on `version=="dev" || output != "human" ||
      BIFROST_CHECK_UPDATE=="false"` (the extra `output` gate is bifrost-specific — heraut has no
      output modes), 500ms timeout, cache under `UserCacheDir/bifrost/update-check.json`, errors
      swallowed. ~25 `NewRootCmd()` test callers updated to `NewRootCmd("dev")`. 146 tests green
      incl. `-tags integration`. **Open (app-side, flagged for user):** released binaries stay
      silent until `.goreleaser.yml` ldflags gain `-X main.Version={{.Tag}}` (a CI/CD touch) — see
      the M5.4 distribution task.
**Distribution** *(split after exploration — see the decisions below)*. The two apps'
goreleaser configs had diverged (bifrost: archives + goreleaser-owned release; heraut: raw
versioned binaries + `release: disable`, heraut owns the release per its ADR-0013/0018). The
user's call: **heraut is the family's release tool** (it will release bifrost and future
tools), so **heraut's raw-binary model is canonical** and bifrost converges to it. OSS
GoReleaser has no remote-include (`includes:` is Pro-only), so the shared config is a
**copy-and-adapt template documented in forge**, not a live dependency. Release *workflows*
stay per-app (heraut = self-release + Docker/GHCR, bifrost = self-release only — too divergent
to unify now). The Homebrew tap goes in a **dedicated `adaouat/homebrew-tap` repo** (details
deferred — user's call). **Raw binaries are retained:** ADR-0013's original driver (avoiding
tar/zip extraction + zip-slip in the self-updater) is now historical (self-updater removed in
M5.2/ADR-0005), but raw binaries keep the curl install a one-liner and are consumed directly
by both mise (`github` backend) and Homebrew (goreleaser generates a per-platform raw-binary
formula) — no archive needed.

- [x] **Goreleaser convention (forge docs).** A `docs/guides/` distribution guide + an
      annotated sample `.goreleaser.yml` pinning the canonical raw-binary model (version
      ldflags, `formats: [binary]`, checksums, mise/curl/Homebrew channels, release-ownership
      options). forge stays a pure library — this is a template apps copy, not code. **Done:**
      `docs/guides/distribution.md` (model, why-raw-binaries with the now-historical ADR-0013
      self-update rationale, release ownership, mise/curl/Homebrew channels, what-isn't-shared)
      + `docs/guides/goreleaser.sample.yml` (annotated `<app>` template, validates clean via
      `goreleaser check`). The Homebrew `brews:` skeleton lives in the guide as a fenced block
      (yamlfmt re-indents trailing comments in the `.yml`, which misattached them under
      `checksum:`); the sample shows `release: disable: false` (self-release) as the active
      choice with the heraut-owned variant in a leading comment. Created `docs/guides/` + its
      index, registered in `docs/README.md`.
- [x] **Converge bifrost's `.goreleaser.yml`** to the canonical model (versioned binary name,
      `formats: [binary]` dropping tar/zip). Keep `release` enabled for now — `release: disable`
      waits until heraut-driven release of bifrost is wired. Add bifrost install docs (mise + curl).
      **Done** (bifrost `ci(release): converge to raw-binary asset model` + `docs(readme): add
      install instructions`): dropped the tar.gz/zip archives for `formats: [binary]` + versioned
      `builds.binary`; release stays goreleaser-owned. Created bifrost's README (was empty) with
      go install / mise / curl sections mirroring heraut. Verified via `goreleaser release
      --snapshot`. **Finding — explicit `name_template` required:** `format: binary`'s default
      name_template keys off `.Binary`, so a versioned `builds.binary` doubles the
      version/os/arch (`bifrost_<v>_<os>_<arch>_<v>_<os>_<arch>`) on goreleaser **< 2.16**. The
      fix is an explicit `name_template` using `.ProjectName` — which both the forge sample **and
      heraut** already carry (a first pass mistakenly dropped it for bifrost; the snapshot caught
      the doubling and it was restored). bifrost now matches heraut + the sample. **Follow-ups
      (all fixed this session):** (a) goreleaser pin skew — bifrost bumped `2.15`→`2.16` to match
      heraut (`build(mise)`); both CIs float `~> v2`, so released assets were never affected;
      (b) bifrost now gitignores `dist/` (`chore`); (c) heraut README cleared of the five stale
      `self-update` references left by M5.2 (`docs`). *(An earlier note here claimed heraut's
      goreleaser was fragile — incorrect; heraut already had the explicit name_template.)*
- [x] **Homebrew tap + cask blocks.** **Done:** a shared `adaouat/homebrew-tap` repo (README +
      MIT license + `Casks/`); both apps wired with `homebrew_casks` (not the deprecated `brews` —
      Homebrew prefers casks for pre-built binaries) and `builds.binary` made plain `<app>` so the
      cask installs under that name (asset name kept versioned via `name_template`). bifrost is the
      clean case (goreleaser-owned release → default URL, goreleaser pushes the cask); **heraut**
      (build-only, `release: disable`) needed an explicit `url.template` + a post-release push step
      (skips gracefully without the token) + an `artifacts.json`-based collect rewrite (since plain
      `builds.binary` leaves build outputs unversioned on disk). All snapshot-validated; forge's
      distribution guide + sample updated to the cask recipe. **Pending (user, GitHub-side):**
      create the tap repo on GitHub (public) + push, the fine-grained `HOMEBREW_TAP_TOKEN` PAT/org
      secret, then the first release publishes each cask. ⚠️ bifrost's release *fails* without the
      token (goreleaser-owned push); heraut's skips gracefully.
- [x] **Shared lint/CI reusable workflow** *(separate track, not release-related — surfaced
      during M5.4)*. A `workflow_call` workflow **hosted in forge** ([ADR-0006](../adr/0006-shared-ci-reusable-workflow.md)),
      called by bifrost + heraut **and forge itself** (3 consumers). Scope: **lint + test only**
      (golangci-lint, govulncheck, `go test`, coverage gate) — *not* build/release/Docker, which
      stay per-app. **Done:** ADR-0006 + `forge/.github/workflows/go-ci.yml` (reusable lint +
      test, single **required** `coverage-threshold` input — no default, so forge never silently
      governs a caller's gate; per-project policy, bifrost 20 / heraut & forge 85); forge's own `ci.yml` dogfoods
      it via `uses: ./.github/workflows/go-ci.yml`, and bifrost + heraut now call
      `adaouat/forge/.github/workflows/go-ci.yml@79edf69` (**v0.6.0**, SHA-pinned per the rule) —
      bifrost passes `coverage-threshold: 20`, heraut uses the default 85; both keep their build
      jobs inline. Scope narrowed in implementation: the **only** cross-repo difference is the
      coverage threshold — CI runs no integration-tagged tests anywhere (they sit behind a build
      tag, excluded by plain `go test ./...`), so no `integration` input was needed. The three
      caller workflows pass `actionlint`. *(Note: forge is already public **and released at
      v0.6.0** via cocogitto — overtaking the M6 "untagged until v0.1.0" assumption; M6 below
      needs reconciling.)*

## M6 — Finalize (depend on the published forge)

*Reconciled: forge is already public and released via cocogitto (now **v0.6.2**), overtaking the
original "cut v0.1.0" plan — "tag forge" is done. The remaining finalize work:*

- [x] **Depend on the published forge.** Drop the apps' `replace github.com/adaouat/forge =>
      ../forge`, set the `github.com/adaouat/forge` require to the published tag (**v0.6.2**),
      `go mod tidy` both, and verify build + tests stay green. (forge itself has no replace.)
      **Done:** both apps now require `forge v0.6.2` with no replace. Note: `go get @v0.6.2` failed
      because dropping the replace left an unresolvable `v0.0.0` in the graph, so used
      `go mod edit -dropreplace` + `-require=…@v0.6.2` + `go mod tidy` instead. bifrost build
      (incl. `-tags integration`) + heraut build + both full test suites green; `go mod tidy -diff`
      clean. The `replace`-directive era (since M1.3) is over — forge is a normally-versioned dep.
- [x] Per-package contract ADRs in `docs/adr/` (one per Tier-1 package whose interface is
      now load-bearing across two repos). **Done as a consolidated ADR** —
      [ADR-0007](../adr/0007-public-api-surface-and-stability.md) enumerates every Tier-1
      package's load-bearing exported surface + its governing ADR (0002 exec, 0003 exitcode,
      0004 ui, 0005 updatecheck; `config`'s contract fixed in 0007 itself) + the stability
      commitment (breaking change → new ADR + coordinated bump). Chose one ADR + references over
      six near-duplicates of 0002–0005, per forge's YAGNI ethos.
- [x] Document the Tier-2 sync workflow (how an app refreshes its `.claude/rules` /
      `.config` from forge's canonical `docs/rules` / `.config`) in `docs/guides/`. **Done:**
      `docs/guides/tier2-sync.md` — what forge canonicalizes (rules + `.config` baseline) vs
      app-owned, the deliberate diff-and-apply process (sibling `diff -u`, minding the
      `agent.md`↔`claude.md` rename), and the direction (forge upstream; promote-then-sync) plus
      cadence. Registered in the guides index.

**M6 complete — and with it the M0–M6 roadmap.** forge is a published library (`v0.6.2`)
consumed by bifrost + heraut off the tag (no `replace`); its public contract is pinned
(ADR-0007) and the Tier-2 sync is documented.

## M7 — Family UI theme *(post-roadmap)*

*Surfaced after M6: a cohesive `fang` theme for the family (in the spirit of glab's), shared via
forge `ui`. Accents chosen: bifrost **Aurora** (teal/violet), heraut **Heraldic** (gold/azure).*

- [x] forge **`ui.Palette`** — the shared structural colors (text, muted, dim, argument,
      success/warn/error), light/dark adaptive, semantic colors matching the existing status
      helpers. forge stays **framework-agnostic** (no `fang`/`cobra` dep): apps assemble their
      `fang.ColorScheme` from the palette + a per-tool accent. [ADR-0008](../adr/0008-ui-theme-palette.md).
      **Done** (forge `feat(ui): add Palette …`, released **v0.7.0**): `ui.Palette` +
      `NewPalette(lipgloss.LightDarkFunc)`, TDD (light/dark table), lint clean.
- [x] Wire **bifrost** (Aurora: teal accent / violet secondary) + **heraut** (Heraldic: gold /
      azure): a `fang.WithColorSchemeFunc` built from `ui.Palette` + the app's accent. **Done:**
      both apps bumped to forge v0.7.0; each has a `cmd/<app>/theme.go` `colorScheme(c)` mapping
      `ui.NewPalette(c)` + accent → `fang.ColorScheme`, wired via `fang.WithColorSchemeFunc` in
      `main`. Snapshot-verified: themed `--help` (teal titles for bifrost, gold for heraut);
      builds + suites + lint green.
- [x] *(refinement)* align `ui` status/spinner colors to the palette so status output matches the
      theme end-to-end. *(Most status colors matched the palette's semantic values, but `Info` used
      `#6B7280` and the spinner used ANSI `"214"`; route them through one source.)* **Done:** all color
      literals now live once in `palette.go`; the three semantic colors (`colorSuccess`/`colorWarn`/
      `colorError`) back both `Palette.Success/Warn/Error` and the status helpers, so they can't drift.
      `Info` now references the palette's muted neutral (the light/darker variant — legible on both
      backgrounds) instead of an orphan `#6B7280`, and the spinner glyph is hex (`214` → `#FFAF00`).
      The helpers stay fixed-color: their `io.Writer` signature (ADR-0007) carries no light/dark
      context, so *adaptive* routing (Info tracking the background-resolved muted) would need a
      breaking signature change — out of scope. The apps' ASCII-art banner (renders in `Base`) is an
      app-side render change, still deferred. Behaviour-preserving except Info's gray nudges
      `#6B7280` → `#6E7781` (imperceptible); `ui` suite (48 tests, semantic hexes pinned) + lint green.
- [x] *(fix)* fang `--help` USAGE block rendered `[command] [--flags]` invisibly (gray-on-gray). fang
      uses the `Codeblock` slot as the usage block's *background* (`Background(cs.Codeblock)`) and draws
      the placeholders with `DimmedArgument`; forge mapped `Codeblock: p.Muted` — a mid-gray *foreground*
      color — so `p.Dim` text sat on a `p.Muted` background (~1.45:1). **Done:** added `Palette.Surface`
      (a subtle elevated background, `#EAEEF2` light / `#22272E` dark) and remapped `Codeblock: p.Surface`
      in `ui.ColorScheme`, so placeholders read on the block while the program name stays accent-colored.
      Additive (new palette field, no API break); bifrost + heraut get the fix on their next forge bump.
      TDD (palette `Surface` + ColorScheme `Codeblock`); `ui` suite + lint green.
- [x] *(refinement)* `config.Load` on an empty file returns the raw `config: EOF` (`config/loader.go`
      — yaml's `io.EOF` wrapped verbatim). Map `io.EOF` to a clearer "empty config" message, or treat
      an empty file as a zero-value config. Low priority; the `config` contract is fixed in
      [ADR-0007](../adr/0007-public-api-surface-and-stability.md). **Done:** added the sentinel
      `ErrEmptyConfig` (errors.Is-classifiable, not a string-matched "EOF"), returned when Decode hits
      `io.EOF`; recorded in the ADR-0007 surface table. TDD (empty reader + empty file). Lint + suite green.
- [x] *(refinement)* cap the GitHub response body with an `io.LimitReader` before decoding
      (`updatecheck/check.go`). The API is trusted and bounded, so risk is negligible — but a
      defensive cap is cheap insurance for a library. **Done:** `maxResponseBytes = 1<<20` cap via
      `io.LimitReader` in `Checker.latest`; TDD (an over-cap body now fails to decode rather than
      being read unbounded). Lint + suite green.

- [x] **Shared release setup** — composite action ([ADR-0009](../adr/0009-release-setup-composite-action.md)).
      The three release.yml share an identical prelude (mise → install heraut → GPG → identity →
      resolve version); extract it as `forge/.github/actions/release-setup` (a **composite action**,
      not a `workflow_call` workflow — the setup is a step-prefix the apps' build/release continue in
      the *same* job, which a separate-job reusable workflow can't do). build/cask/Docker stay
      per-app. **Partial (forge-side done):** ADR-0009 + `release-setup/action.yml` (inputs:
      gpg-private-key/github-token/version; outputs version/tag; also sets `$VERSION` in env) landed,
      and forge's own `release.yml` uses it via `./.github/actions/release-setup` (actionlint-clean).
      **Done:** released in forge **v0.7.2**; bifrost + heraut rewired to
      `adaouat/forge/.github/actions/release-setup@66461c2` (# v0.7.2) — each dropped its five
      setup steps for the one action step (checkout stays). All three actionlint-clean. heraut's
      bootstrap-vs-`$FRESH_BIN` distinction is preserved (the action resolves the version with the
      bootstrap heraut; heraut's own preflight/release still use the freshly built binary).

## M8 — forge as the CLI framework foundation *([ADR-0010](../adr/0010-cli-framework-foundation.md))*

*Reframe forge from framework-agnostic utilities to **the foundation every tool imports** — it
owns the CLI framework layer (fang, huh, theme), cutting version drift (cobra is already split
1.9.1/1.10.2) and the per-app theme duplication. Zero-domain-logic still holds. Supersedes ADR-0008.*

- [x] **ADR-0010** — record the identity shift (forge owns fang/huh/theme; cobra aligned not
      wrapped; viper dropped; zero domain logic unchanged).
- [x] **Align cobra** — bump bifrost `1.9.1` → `1.10.2` (heraut's pin); document the family pin in
      the Tier-2 baseline. *(Independent of the rest — fixes the live drift now.)* **Done:** bifrost
      `build(deps): align cobra to 1.10.2` (+ pflag 1.0.9), build + suite green. Both apps now on
      cobra 1.10.2; the Tier-2-baseline note rides along when forge gains the dep via fang.
- [x] **forge `cli.Run`** — wrap `fang.Execute` (version + theme); forge gains fang. Apps drop their
      direct fang import. TDD. **Done (with the theme, below — they're coupled):**
      `cli.Run(ctx, cmd, version, accent)` wraps `fang.Execute`; forge now requires fang 2.0.1 +
      cobra 1.10.2. Apps will drop fang when they adopt it (after the M8 release). Smoke-tested.
- [x] **forge theme** — `ui.ColorScheme` (default accent + per-tool override); the fang mapping
      moves into forge. **Done:** `ui.Accent` + `ui.ColorScheme(ld, accent)` — the slot mapping
      (palette → fang.ColorScheme) lives once in forge, TDD'd. **Default accent: Ember** — forge's
      brand (orange `#EA580C`/`#FB923C` over coal-red `#BE123C`/`#FB7185`), the forge fire;
      `ui.DefaultAccent()` exposes it and a zero `Accent` falls back to it, so a new tool gets a
      theme without picking one.
- [x] **forge owns huh** — move the huh dep + any shared huh helpers/theme; apps drop direct huh import.
      **Done (refined):** forge gained huh 2.0.3 + `ui.HuhTheme(accent)` — a `huh.ThemeFunc` (from
      `ThemeBase`) branding the focused state with the accent + errors from the palette, TDD'd.
      Correction to the original framing: the apps **keep** their direct huh import (they build
      their own forms — huh is aligned-by-convention like cobra, not dropped); they wire
      `form.WithTheme(ui.HuhTheme(accent))` when they adopt.
- [x] **Apps adopt** — bifrost + heraut use `cli.Run` + `ui.ColorScheme`; delete `cmd/<app>/theme.go`;
      re-pin to the M8 forge release. **Done (forge v0.8.0):** both apps call
      `cli.Run(ctx, cmd, version, ui.Accent())` (accent in each app's `tui`/`ui` package — bifrost
      Aurora, heraut Heraldic), deleted their `cmd/<app>/theme.go`, and **fang is now `// indirect`**
      (via forge). huh prompts are branded too via `ui.HuhTheme()` — bifrost's 4 standalone fields
      inline, heraut's 8 wizard forms through a `themedForm` helper. Themed `--help` verified (teal /
      gold); suites + lint green.
- [x] **`docs/guides/new-tool.md`** — "start a tool on forge" (the batteries-included path).
      **Done:** the guide walks a new CLI from `go mod init` → `cli.Run` + an accent → Tier-2
      scaffolding → the shared `go-ci.yml` → heraut-driven release via `release-setup` → the
      update hint, linking distribution.md / tier2-sync.md and ADRs 0006/0007/0009/0010 rather
      than duplicating them. Registered in the guides index.

**M8 complete.** forge is now the CLI framework foundation: it owns fang + cobra + huh + the
theme; `cli.Run` + `ui.ColorScheme`/`ui.HuhTheme` drive both apps (fang dropped to indirect); the
Ember default + per-tool accents are in place; cobra drift fixed; and a new tool starts from one
guide.

---

## M9 — Logging foundation *([ADR-0011](../adr/0011-logging-foundation.md))*

*Same shape as M8: a mechanical CLI-runtime concern (verbosity → level mapping, output
handler, TTY-aware rendering) that's identical across the family and otherwise drifts per
app, the way fang/huh/the theme did before ADR-0010. Logging *content* (what gets logged,
structured error context, routing to external sinks) stays out — domain logic, per ADR-0001.*

- [x] **Confirm the `charmbracelet/log` import path** — verify a `charm.land` vanity path
      resolves (mirrors the M3 `colorprofile` flag); if not, document the
      `github.com/charmbracelet/log` exception in `docs/rules/coding.md` alongside
      `colorprofile`/`x/term`. **Done:** the unversioned `charm.land/log` *looks* like it
      resolves in the proxy listing but fails on `go get` ("module declares its path as
      `github.com/charmbracelet/log`") — a v1/v0 artifact. Per the project's own
      [UPGRADE_GUIDE_V2](https://github.com/charmbracelet/log/blob/main/UPGRADE_GUIDE_V2.md),
      the v2 line is properly republished: `charm.land/log/v2` resolves and `go get`s cleanly
      (verified — added as a real `require`, no error). **No exception needed**; the M9
      logging-setup package imports `charm.land/log/v2`, matching the `huh`/`bubbles`
      `/v2` convention already in use.
- [x] **forge logging setup** — a thin package wrapping `log/slog` (the API) with
      `charm.land/log/v2` as the rendering `slog.Handler` (the backend), in the same
      interface/implementation relationship as `cli.Run`/fang. Wires `--verbose`/`--quiet`
      to levels, routes to stderr, and respects `ui.IsTTY`/`ui.HasColor` for plain-vs-colored
      output. Exposes a constructor returning a `*slog.Logger` — exact name/signature TBD at
      implementation time (TDD: failing test first, per `docs/rules/testing.md`). **Done:**
      `log.New(w io.Writer, level slog.Level) *slog.Logger` (`log/log.go`) — genuinely thin:
      `slog.New(charmlog.NewWithOptions(w, charmlog.Options{Level: charmlog.Level(level)}))`.
      **Correction to the framing:** no separate `ui.IsTTY`/`ui.HasColor` wiring needed —
      `charm.land/log/v2` already runs `colorprofile.Detect(w, os.Environ())` internally (the
      same mechanism `ui.HasColor` wraps), so duplicating it would be redundant. `slog.Level`
      and `charmlog.Level` share identical numeric values for Debug/Info/Warn/Error
      (-4/0/4/8), so the conversion is a direct `charmlog.Level(level)` — no lookup table.
      Verbosity-flag → level mapping and stderr routing are the **app's** call at the
      `cli.Run`/command-construction site (forge hands back a configurable `*slog.Logger`,
      it doesn't own the flags). TDD: table-driven `TestNew` (level filtering: info logger
      reports info+ filters debug; warn logger reports warn+ filters info) +
      `TestNew_writesToProvidedWriter`, both via a `bytes.Buffer`. `go mod tidy` promoted
      `charm.land/log/v2` to a direct `require`. Lint + suite green (133 tests, 8 packages).
- [ ] **Apps adopt** — bifrost + heraut migrate any ad hoc logging (`fmt.Println`/`log.Printf`)
      to the shared `*slog.Logger`; re-pin to the M9 forge release.
  - **bifrost — deferred.** Investigated first: bifrost has **no ad hoc logging to migrate**.
    Its only logging-related code is one line — `clog.SetLevel(clog.DebugLevel)` in the
    `--verbose` path (`internal/cmd/root.go`), toggling `charm.land/log/v2`'s *global* logger.
    There are zero `clog.Info/Debug/Error` call sites and no logger construction anywhere. forge's
    `log.New` returns a `*slog.Logger` *instance* (deliberately no global setter), so there's
    nothing for bifrost to wire it into — constructing one would be an unused logger (YAGNI, which
    forge's coding rules forbid). Deferred until bifrost grows real logging needs; bifrost is left
    untouched (the dead/global `SetLevel` line is bifrost's own call to revisit, not part of M9).
    The `log` package still ships in the M9 forge release so it's available when a consumer
    appears (heraut next).

## Explicitly NOT on this roadmap

Per ADR-0001 Tier 3: config **schemas** and **merge semantics**, bifrost's hook runner and
atomic strategy, heraut's pipeline / generators / platforms / versioning. If a future need
makes one of these genuinely shared, it earns its own ADR first — it does not get bolted on
here.
