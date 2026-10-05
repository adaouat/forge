# ADR-0016 — `config.Resolver` expands a leading `~`

**Status:** Accepted
**Date:** 2026-10-05

## Context

A consumer CLI set its config path in a mise config as a quoted value,
`<APP>_FILE = "~/.config/<app>.yml"`. A shell expands `~` only in an *unquoted* assignment; a
quoted mise `[env]` value, a `.env` file or a launchd plist hands the process a literal `~`.
`Resolver.Resolve` returned the `--config` value and `<APP>_FILE` untouched, so the tool opened
the relative path `~/.config/<app>.yml` and failed with `no such file or directory`.

The defect sits in the shared primitive: heraut (`HERAUT_FILE`) and bifrost (`BIFROST_FILE`)
resolve through the same `Resolver` and have the same bug. A per-app wrapper would be the
identical few lines in every consumer, so the fix clears the [ADR-0001](0001-shared-core-module.md)
bar (identical, stable, ≥2 consumers, no domain logic) and belongs in forge.

## Decision

1. Add `config.ExpandHome(path string) string`. `"~"` and paths starting with `"~/"` become the
   user's home directory (`os.UserHomeDir`) joined with the rest; every other path is returned
   as is.
2. `Resolver.Resolve` expands both of its inputs: `explicit` before the precedence logic, and
   the path it returns (which covers the `<APP>_FILE` value `Resolve` reads itself). The
   returned `Source` is unchanged. `Label`, `InitDest`, `Load` and `Decode` keep their behaviour.
3. **`~user` is deliberately not supported.** `~user/x`, `~name.yml` and a `~` anywhere but the
   start are returned unchanged. Looking up another user's home needs `os/user` (cgo or
   `/etc/passwd` parsing, platform-dependent) for a case no config path needs.
4. **No home directory → unchanged.** If `os.UserHomeDir` fails or returns `""`, `ExpandHome`
   returns the path as written instead of an error, so the later open failure names exactly
   what the user typed. `ExpandHome` never panics.

## ADR-0007 classification

`ExpandHome` is **additive** (a new exported symbol). The `Resolve` change is a change of
**documented behaviour**: it used to return the flag/env value verbatim, so a path starting
with `~/` meant a relative directory literally named `~`. Per
[ADR-0007](0007-public-api-surface-and-stability.md) this is treated as a behaviour change,
hence this ADR.

It is safe in practice. A directory literally named `~` in the working directory is
practically never intended. It is the classic result of exactly this quoting bug, and every
real `"~/..."` value was already failing. Both consumers call `Resolve` without inspecting
the path for `~` (heraut `internal/config/path.go`, bifrost `internal/cmd/cmdutil/path.go`),
and neither needs a code change. As with [ADR-0013](0013-exitcode-summary-and-full-error.md),
the "coordinated bump" ADR-0007 asks for is a plain `go get` of the new tag in each consumer,
with no API migration. It can ride their next routine forge bump rather than being forced
immediately.

## Consequences

- `--config ~/x.yml` from a non-shell launcher, and a quoted `~/...` in `<APP>_FILE`, now
  resolve to the home directory in every consumer after the bump.
- bifrost's `ResolveInitDest` reads `BIFROST_FILE` itself rather than via `Resolve`, so
  `bifrost init` does not get the expansion from this change. That is a bifrost-side
  follow-up (`config.ExpandHome` on that value).
- bifrost also carries its own `expandHome` for SSH key paths (`internal/transport/ssh.go`),
  which returns an error instead of leaving the path unchanged. It can adopt
  `config.ExpandHome` later if those semantics fit; nothing here forces it.
- ADR-0007's `config` row gains `ExpandHome` and this ADR as its governing reference.
