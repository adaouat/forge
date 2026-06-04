# ADR-0008 — Family UI theme: shared palette, per-tool accent

**Status:** Accepted
**Date:** 2026-06-04

## Context

The family wants a cohesive CLI theme (in the spirit of glab's GitLab theme). `fang` colorizes
help, version, and errors via a `fang.ColorScheme` (≈15 slots), resolved per terminal background
with a `lipgloss.LightDarkFunc`. Two forces pull apart: a **shared family look** (structure,
neutrals, semantic colors) and a **per-tool identity** — bifrost = Aurora (teal/violet), heraut =
Heraldic (gold/azure). glab bakes in one brand; this family has several tools.

## Decision

forge `ui` owns the **shared structural palette**; each app owns its **accent** and assembles the
`fang.ColorScheme`.

- **`ui.Palette`** — `Text`, `Muted`, `Dim`, `Argument`, `Success`, `Warn`, `Error`
  (`color.Color`), resolved by `NewPalette(lipgloss.LightDarkFunc)`. The semantic colors
  (success/warn/error) match the existing `ui` status helpers, so status output and the fang theme
  agree.
- **forge stays framework-agnostic.** `ui.Palette` does **not** import `fang`/`cobra`. forge sits
  *below* the CLI framework (the apps own the cobra/fang layer); pulling fang + cobra into a
  deliberately lean library (cf. [ADR-0005](0005-updates-via-package-managers.md)) just for a color
  struct would breach that separation. The app assembles the `fang.ColorScheme` — a ~15-line,
  framework-specific mapping — from the palette + its accent, the same kind of per-app glue as its
  `cobra.Command` wiring.
- **Per-tool accent** (the brand) lives in the app: `Title`/`Program`/`Flag` ← accent, `Command` ←
  secondary, everything else ← the shared palette. bifrost Aurora (teal `#0E8A8A`/`#2DD4BF`, violet
  `#8250DF`/`#D2A8FF`); heraut Heraldic (gold `#9E6A03`/`#E3B341`, azure `#0969DA`/`#79C0FF`).

## Consequences

- One source for the family's structural colors; each tool reads as the family yet is instantly
  distinguishable by accent. A third tool picks an accent and inherits the rest.
- forge gains **no** new dependency; the small fang mapping is duplicated across apps (≤15 trivial
  lines) — accepted per forge's YAGNI ethos, the price of keeping forge framework-agnostic.
- The palette is additive: existing `ui` status/spinner colors already match its semantic colors;
  aligning them to read from `Palette` is an optional follow-up, not required for the theme.
