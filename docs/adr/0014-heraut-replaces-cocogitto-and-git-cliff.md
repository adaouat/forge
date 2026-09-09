# ADR-0014 — `heraut` replaces `cocogitto`/`git-cliff` in the canonical tool stack

**Status:** Accepted
**Date:** 2026-09-09

## Context

[ADR-0001](0001-shared-core-module.md) canonicalized `.config/{mise,hk,cocogitto,typos,yamlfmt}`
as Tier-2 scaffolding, inherited unchanged from bifrost/heraut's original stack: `cocogitto`
(`cog`) validated commit messages (`hk`'s `commit-msg` hook) and, with `git-cliff`, generated
release changelogs ([ADR-0009](0009-release-setup-composite-action.md), ADR-0012's "changelog
content and the release process that writes it" line).

heraut has since absorbed both jobs natively as first-class subcommands, and dropped its own
dependency on the external binaries in the process:

- **Commit linting** — `heraut commit verify`/`heraut commit check` parse and validate the
  conventional-commit grammar directly (`internal/cmd/commit.go`), printing a
  cocogitto-style recap. heraut's own `.config/hk/config.pkl` already dogfoods `go run
  ./cmd/heraut commit verify --file {{ commit_msg_file }}` in its `commit-msg` hook and carries
  no `cocogitto` tool, config, or shell alias.
- **Changelog generation** — `internal/generators/native` reimplements git-cliff's grouping,
  tag-walk, and taxonomy behavior (see its inline comments: "the native equivalent of
  git-cliff's `--tag-pattern` regex", "honours git's own topological tag ordering, matching
  git-cliff's tag walk") without shelling out to either binary. The only remaining
  `cog`/`git-cliff` references in heraut are backward-compat config-migration handling for
  apps still on the old external-generator config shape, and test fixtures asserting parity
  with the old output.

Since [ADR-0009](0009-release-setup-composite-action.md) already made `heraut` the shared
release orchestrator for all three repos, this isn't heraut *adding* a new job — it's heraut
finishing the replacement of two external Rust binaries it used to shell out to, with the
side effect that those binaries are no longer part of the family's required toolchain at all.

forge's own `.config` had not caught up: `cocogitto` was still pinned in `mise/config.toml`,
`.config/cocogitto/config.toml` still existed, and `hk`'s `commit-msg` hook still called `cog`
directly — none of it wrong, just stale relative to what heraut (the tool forge's own scaffolding
points consumers at) now does on its own.

## Decision

Forge's canonical Tier-2 scaffolding drops `cocogitto` and `git-cliff` and adopts `heraut`
itself as a `mise` tool (`"github:adaouat/heraut" = "0"`), calling its subcommands directly
from `hk`:

- `mise/config.toml`: remove `cocogitto`/`git-cliff` tool pins and the now-dead `cog` shell
  alias; add `"github:adaouat/heraut" = "0"` (heraut is already installed as a real binary
  here — bifrost/heraut build it from source instead, since it's their own repo).
- `hk/config.pkl`: `commit-msg`'s `cocogitto` step becomes `heraut:commit:verify` (`heraut
  commit verify --file {{ commit_msg_file }}`); a new `heraut:check:config` step
  (`heraut check config`, glob-scoped to `.config/heraut.yml`) validates the app config
  itself, matching what `heraut check` already does at runtime.
- `.config/cocogitto/config.toml` is deleted — there is nothing left that reads it.

**Rollout is per-repo, not atomic.** bifrost still runs `cocogitto`/`cog` today (verified:
`bifrost/.config/mise/config.toml` and `.config/hk/config.pkl` are unchanged) and is not
touched by this ADR — its own migration is a separate, later sync
([tier2-sync.md](../guides/tier2-sync.md)) once this canonical baseline is settled here.

## Consequences

- **One fewer pair of Rust binaries in the required toolchain.** `cocogitto` and `git-cliff`
  are no longer installed by `mise` for any repo that adopts this baseline; `heraut` was
  already a required tool for releases, so this is a net reduction, not an addition.
- **`.config/{mise,hk,typos,yamlfmt}` is the new Tier-2 scaffolding list** — `cocogitto` drops
  out of [ADR-0001](0001-shared-core-module.md)'s "In scope — scaffolding" enumeration and
  every doc that repeated it (`CLAUDE.md`, `docs/guides/new-tool.md`,
  `docs/guides/tier2-sync.md`, `.claude/skills/new-tool/SKILL.md`), updated alongside this ADR.
- **Commit-message linting and changelog content stay app-side in spirit, not just in
  practice**: previously "app-side" meant "cocogitto's config, owned per-app"; now it means
  "heraut's own logic", but the boundary ADR-0001/ADR-0012 drew — forge does not own commit
  grammar or changelog content — is unchanged. heraut is a consumer of forge, not forge itself;
  nothing here crosses the Tier-3 "false friend" line.
- **Follow-up, not required by this ADR**: bifrost (and any future tool built from this
  baseline) still needs the same swap applied to its own `.config`, via the normal
  [tier2-sync](../guides/tier2-sync.md) diff-and-apply process.
