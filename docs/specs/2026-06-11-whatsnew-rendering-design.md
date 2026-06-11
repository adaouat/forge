# `whatsnew` rendering refinements: drop terminal-query styling, add a pager

**Date:** 2026-06-11
**Status:** Draft (pending review)
**Relates to:** [ADR-0012](../../adr/0012-whatsnew-changelog.md)

## Problem

Two issues with `updatecheck.WhatsNewCommand`'s output, both in `render()`
(`updatecheck/whatsnew.go`):

1. **Styling silently degrades to raw markdown.** `render()` calls
   `glamour.Render(md, "auto")`. The `"auto"` style makes glamour query the terminal for its
   background color via OSC11 escape sequences. When the terminal doesn't answer (or answers
   too slowly), glamour times out and `render()`'s fallback (`out = md`) writes **raw,
   unrendered markdown** — literal `#`, `**`, `-`, etc. Reproduced empirically: running
   heraut's `whatsnew` under a pty (`script`) shows 4 OSC11 query attempts followed by raw
   markdown output. This is the "black/white, missing styling" behavior reported.

2. **No pager.** ADR-0012 explicitly deferred a pager ("Add one only if changelog length
   warrants it"). `glab whatsnew` pages its output. With Tier D (multi-release spans +
   embedded changelog fallback), output can be long enough to scroll past the terminal.

## Decision

### 1. Style selection without terminal queries

Replace the `"auto"` lookup with a selection based on `ui.HasColor(w)` (forge's existing
`ui` package — handles `NO_COLOR`, `CLICOLOR_FORCE`, `TERM=dumb`, and TTY-ness with **no**
terminal round-trip):

- `ui.HasColor(w)` → glamour style `"dark"`.
- otherwise → glamour style `"notty"` (markdown-aware formatting, no ANSI codes — strictly
  better than today's raw-markdown fallback).

`updatecheck` gains an internal dependency on the sibling `ui` package. Both packages live
in the forge module; `ui` does not import `updatecheck`, so no cycle.

We always default to a dark-themed render when color is available — most terminals devs use
default to dark themes, and a dark-style render on a light background is still readable
(unlike today's all-or-nothing raw-markdown fallback). No light/dark heuristic beyond that;
no new config surface.

### 2. Pager via `$PAGER`

After rendering, if `ui.IsTTY(w)` is true, pipe the rendered output through a pager rather
than writing it directly to `w`:

- **Pager resolution:** `$PAGER` env var if set and found via `exec.LookPath`, else `less`
  if found via `exec.LookPath`, else no pager (write directly).
- **`less` defaults:** if the resolved pager is `less` (explicit `$PAGER=less` or the
  default) and `$LESS` is unset, set `LESS=FRX` for the child process — git's convention:
  `-F` exit immediately if content fits one screen, `-R` pass through ANSI color codes (the
  `"dark"` glamour style emits these), `-X` don't clear the screen on exit.
- **Opt-out:** `$NO_PAGER` set (any non-empty value) skips paging entirely, matching the
  git/gh convention. `$PAGER=cat` or `$PAGER=""` also effectively disable paging (cat just
  passes content through).
- **Non-TTY `w`** (piped/redirected output, the existing test/automation path): no paging,
  unchanged from today.
- The pager process is a passthrough: rendered markdown → pager's stdin; pager's
  stdout/stderr/stdin connect directly to `os.Stdout`/`os.Stderr`/`os.Stdin` so scrolling
  works. This is **not** routed through `exec.Runner` — Runner buffers
  stdout/stderr, which would defeat an interactive pager. A small dedicated helper in
  `updatecheck` spawns it directly via `os/exec`.

## Architecture / data flow

```
cfg.run()
  → assemble(rels) string         (unchanged — deterministic markdown)
  → render(w, md)
      → style := "notty"; if ui.HasColor(w) { style = "dark" }
      → out, err := glamour.Render(md, style)
      → if err != nil { out = md }   (existing fallback, now rarely hit)
      → if ui.IsTTY(w) && pager available && !noPager:
            runPager(pagerCmd, pagerArgs, out, os.Stdout, os.Stderr, os.Stdin)
        else:
            io.WriteString(w, out)
```

`assemble` is untouched. The seam stays: `render` is the only function touched, plus one new
small helper (pager resolution + spawn) in the same file or a new `pager.go` in
`updatecheck`.

## Error handling

- Pager spawn failure (binary vanishes between `LookPath` and `Run`, etc.) → fall back to
  writing directly to `w`. Paging is a UX nicety; it must never cause `whatsnew` to fail or
  swallow output.
- Existing glamour-error fallback (`out = md`) is preserved as a final safety net, now
  expected to fire only on genuine glamour bugs, not terminal-detection timeouts.

## Testing

- **Style selection**: table-driven test on the `style(w io.Writer) string` helper —
  `ui.HasColor` true/false (using a buffer vs. a `*os.File`/pty-like writer, or by
  setting `NO_COLOR`) → `"dark"` / `"notty"`.
- **Pager resolution**: table-driven test on the pure decision function — `$PAGER` set/unset,
  `$NO_PAGER` set, binary present/absent (via a fake `PATH` with a stub `less`/`cat`
  executable in `t.TempDir()`), `$LESS` set/unset → resolved command + args + env.
- **Non-TTY path** (existing tests use a `bytes.Buffer`, which is never a TTY): confirms no
  pager is invoked, output unchanged — covers the existing CI/test path.
- Actually spawning a real pager against a pty is **not** unit-tested (consistent with
  `exectest.FakeBin` being a "sparingly" tool per `docs/rules/testing.md`); covered by manual
  verification (`docs/rules` `verify` skill / running the built binary under a real
  terminal).

## Documentation updates

- **ADR-0012**: add a "Refinement (decided 2026-06-11)" section — following the existing
  refinement-note pattern already in the ADR — documenting (a) the OSC11 auto-detection
  failure mode observed and the move to `ui.HasColor`-based style selection, and (b) lifting
  the pager deferral with the `$PAGER`/`$NO_PAGER`/`less` design above.
- **Roadmap**: new milestone **M12 — `whatsnew` rendering refinements**, two tasks:
  1. Style selection via `ui.HasColor` (drop `"auto"`).
  2. Pager support via `$PAGER`/`$NO_PAGER`.

  Each task follows the two-step roadmap flow and TDD, one per session per
  `docs/rules/agent.md`.

## Out of scope

- No new CLI flags (`--no-pager`, `--style`, etc.) — env vars only, matching git/gh
  conventions and avoiding new exported surface.
- No light-theme heuristic beyond `ui.HasColor` (e.g. `$COLORFGBG` parsing) — deferred
  unless it proves necessary in practice.
- No built-in TUI pager (bubbletea/viewport) — `$PAGER`/`less` shelling matches existing
  forge dependency footprint (no new deps) and CLI conventions (`git`, `gh`, `glab`).
