# ADR-0012 — `whatsnew`: in-terminal changelog after the update check

**Status:** Accepted
**Date:** 2026-06-10

## Context

forge owns the first half of the update lifecycle ([ADR-0005](0005-updates-via-package-managers.md)):
the `updatecheck` package does the version check (`CheckNewer`) and prints an install-method-aware,
24h-cached upgrade `Hint` — "is there a newer version, and how do I get it" — swallowing every error
so it never breaks a run. It deliberately does *not* replace the binary.

The gap: the hint says a newer version exists and how to install it, but not **what changed**. Today
the only changelog pointer is the releases URL emitted in the `Hinter` *fallback* branch — and only
when no package-manager upgrade command is detected. `glab` closes the same gap with `glab whatsnew`:
it fetches the releases list from the API, filters to versions newer than a stored "last seen", and
renders the release-note markdown in the terminal. It is a natural completion of the lifecycle
`updatecheck` already started.

Two value-props get conflated and must be named, because the source choice turns on them:

- **"What's in the version I'm about to upgrade *to*"** — the nudge case. The new release's notes
  were written *after* the running binary shipped, so **an embedded `CHANGELOG.md` cannot answer
  this** — by construction. This needs the **API**.
- **"What changed in the version I'm *running* / recent history"** — an **embedded** changelog
  answers this, offline and exactly.

The bar question is the same as [ADR-0011](0011-logging-foundation.md)'s: only the *mechanical* part
clears [ADR-0001](0001-shared-core-module.md). Fetching, filtering, rendering, caching, and wiring a
command are identical in shape across `bifrost`/`heraut`/the next tool. What is *in* a changelog —
and the release process that produces it (cocogitto, [ADR-0009](0009-release-setup-composite-action.md))
— is app-side and stays there.

## Decision

forge gains a **`whatsnew` capability as an extension of `updatecheck`** (not a new package), exposed
as a **forge-owned `*cobra.Command` constructor** plus a one-line augmentation of the existing hint.
A forge command is on-precedent: [ADR-0010](0010-cli-framework-foundation.md) already put `cobra`/`fang`
in forge's public surface (`cli.Run` takes a `*cobra.Command`). Per-app input is *configuration*
(repo, bin, cache/state path, and — at D — the embedded changelog), exactly the shape `Hinter`
already has; no app-specific name leaks.

**Final target = Tier D** (hybrid source: cached → live API → embedded changelog, glamour-rendered),
**shipped in tiers** so the simplest value lands first and the heavier machinery is gated on appetite
— this is a low-audience "nerd" feature and the rollout reflects that. The renderer is an isolated
seam (`assemble(rels) string` builds the markdown, a separate `render` writes it), so glamour rides
in with C rather than waiting for D: C's whole reason to exist over A *is* the richer in-terminal
read, so a plain-text intermediate would be a stepping stone thrown away. The seam keeps `assemble`
deterministically testable regardless of renderer.

- **Tier A — ship first.** Augment `Hinter.Print` to *always* append a changelog pointer
  (`… available — run: <cmd> · what's new: <releases URL>`), regardless of which upgrade branch
  fires. No new command, no dependency, ~2 lines. Independently worth it and the keep-it-simple
  floor: the changelog is one click away even if C/D never ship.

- **Tier C — the `whatsnew` command, API source, glamour-rendered.** `updatecheck.WhatsNewCommand(cfg)`
  returns a wired cobra command. Source = GitHub releases API. The 24h update check (already running)
  is extended to **also cache the latest release `body` (+ `html_url`)** alongside `tag_name`, so the
  common "just got nudged, one version behind" case renders **instantly and offline with no extra
  call**. Cold/missing cache, or several versions behind (the single cached body can't cover the
  span), → one live `GET …/releases` list call, filtered to `> current`. Rendering goes through the
  `assemble`/`render` seam: `assemble` builds the markdown (tested deterministically), `render` runs
  it through **glamour**, falling back to the raw markdown if glamour errors (glab's pattern). The
  cache change is backward-compatible: an old entry with no `body` simply falls through to the live
  path. This tier onboards the glamour dependency (see Consequences).

- **Tier D — final target: embedded offline fallback.** Add the app's embedded `CHANGELOG.md`
  (`go:embed`) as the **offline fallback**, completing the source primacy order: **cached body →
  live API → embedded changelog**. The embed answers "what's in the version I'm running" when offline
  with a cold cache. Rendering is already glamour from C; D adds only the third source, so the apps
  supply the embedded FS and forge slots it in as the last fallback.

**Filter semantics.** The chosen filter is **newer-than-current-version** (what sits between what I
run and what's upstream) — simpler than glab's persisted "last seen" and matching the upgrade-nudge
intent. A persisted last-seen ("only the delta since I last looked") is a deferred refinement, not
required for D.

**Out of scope** (app-side, or not at all):

- **Changelog *content* and the release process that writes it** — cocogitto + the app's release
  flow ([ADR-0009](0009-release-setup-composite-action.md)) own it. forge only fetches/renders.
- **Non-GitHub sources** — GitHub-only, matching `updatecheck`. No GitLab/Gitea SDKs.
- **Auto-running `whatsnew`** — it stays an *explicit* user command; the nudge only *points* at it,
  consistent with ADR-0005's "no surprise actions on the user's binary".
- **A pager** — glab pages its output; we render non-paged first. Add one only if changelog length
  warrants it (deferred, avoids an extra dep).

## Consequences

- **Tier A is essentially free** and lands the 80/20 immediately; C/D are the real work and gated on
  appetite — the design is recorded here so the decision survives even if D waits.
- forge's exported surface grows **at C** (`WhatsNewCommand` + its config struct) and the
  `updatecheck` cache-entry schema gains `body`/`html_url`. [ADR-0007](0007-public-api-surface-and-stability.md)'s
  surface table updates **per-tier as each lands** — not now; nothing is built yet.
- The cache schema change is **backward-compatible** — older `{checked_at, latest}` entries unmarshal
  with an empty body and degrade to the live/embed path. No migration.
- **GitHub API pressure is negligible**: `whatsnew`'s live path is an explicit, low-frequency user
  command, and the cache makes the nudge case a **zero-call** render. Unauthenticated 60 req/hr/IP is
  ample.
- **At C, forge gains glamour** (goldmark + chroma — a few MB, same charm family as the existing `ui`
  stack). The proxy lists `charm.land/glamour/v2 v2.0.0` — the same `/v2` shape as
  `charm.land/log/v2`, which `go get`s cleanly — so the dependency looks low-risk. The definitive
  `go get` check runs at C implementation time (proxy *listing* alone isn't proof — the unversioned
  `charm.land/glamour` lists up to v1.0.0 but, per the ADR-0011 lesson, may still declare
  `github.com/charmbracelet/glamour` and fail); if `/v2` fails, it is a documented
  `github.com/charmbracelet/glamour` exception in `docs/rules/coding.md` alongside
  `colorprofile`/`x/term`. The isolated `render` seam means even a surprise here doesn't block the
  fetch/cache machinery — `render` degrades to raw markdown.
- Does **not** relax the ADR-0001 bar: forge owns fetch/filter/render/cache/command-wiring (identical
  across the family); changelog content and the release process that produces it stay app-side.
- **≥2 consumers**: bifrost + heraut both ship GitHub releases with cocogitto changelogs; the third
  tool inherits the command free, the same way it inherits the hint.

### Phasing → roadmap

The tiers map one-to-one to M10 tasks (`docs/tasks/roadmap.md`): **A** (hint pointer), **C**
(`WhatsNewCommand`, API + cached body, glamour-rendered), **D** (embedded offline fallback). Each
lands in its own session per the one-task-per-session discipline (`docs/rules/agent.md`), TDD first.
