---
name: bump-forge
description: Use when rolling a published forge release out to its consumer CLIs (heraut, bifrost, …) — bumps each app's go.mod to a forge tag, builds/tests/lints, and commits go.mod/go.sum only. Semi-automated: pauses at the commit type/message and never pushes.
---

# Bump the consumer apps to a forge release

Roll a **published** forge tag out to every `github.com/adaouat/*` CLI that depends on it. forge is
released separately (`heraut release` on forge); this skill is the **consumer side** — run it after the
tag is pushed. It does, per app: `go get forge@<version>` → `go mod tidy` → build/test/lint → commit
**`go.mod`/`go.sum` only**.

**Semi-automated.** **Stop and ask** at these two points:

| Judgment point | Why it's yours |
|---|---|
| **Commit type/message** | `fix:` when the forge release carries a *user-visible* fix the app inherits (e.g. the v0.9.0 usage-block fix), so it lands in the app's changelog; `build:` / `chore(deps)` for a purely internal/infra bump. Read forge's `CHANGELOG.md` / release notes for the version. |
| **Push** | Never push automatically — report and let the user push. |

## Consumer tools

| Tool | Path (sibling of forge) |
|---|---|
| heraut | `../heraut` |
| bifrost | `../bifrost` |

When a new adaouat/* CLI is bootstrapped (see the **new-tool** skill), add a row here. To bump a
subset, pass tool names as extra args; default is all of them.

## 0. Inputs & guards

- **Version** — a forge tag like `v0.9.0`. If not given, resolve the latest:
  `git -C ../forge tag --sort=-v:refname | head -1` (re-fetch tags first if unsure).
- **The tag must be pushed**, not just local — the apps fetch it over the network/proxy, not from your
  working copy. Verify: `git -C ../forge ls-remote --tags origin <version>` returns a ref. Abort if not.
- For each app, **`go.mod`/`go.sum` must be clean** before starting (other dirty files are fine — you
  won't stage them). If `go.mod` is already modified, stop and surface it rather than mixing changes.

## 1. Per tool — bump, then verify

Run each command as `cd ../<tool> && <cmd>` in a **single** invocation (cwd doesn't persist between
shell calls). For each tool in the list:

```bash
cd ../<tool> && GOFLAGS=-mod=mod go get github.com/adaouat/forge@<version> && go mod tidy
cd ../<tool> && go build ./...
cd ../<tool> && go test ./...
cd ../<tool> && golangci-lint run ./...   # or: mise run lint:check
```

All green before committing. A red build/test means forge changed a contract the app relied on —
**stop**, report which app and what broke; don't paper over it.

## 2. Commit — `go.mod`/`go.sum` ONLY

```bash
cd ../<tool> && git add go.mod go.sum && git commit -m "<type>: <subject> (bump forge to <version>)"
```

- **Never `git add -A` / `git add .`** — apps routinely carry WIP (e.g. a modified `.goreleaser.yml`)
  and untracked tool dirs (`.codegraph/`). Stage exactly `go.mod go.sum`, nothing else.
- After committing, confirm the scope: `git show --stat HEAD` shows only `go.mod`/`go.sum`, and
  `git status --short` still shows any pre-existing WIP untouched.
- Branch: pre-v1.0 apps commit on `main` (forge convention, inherited). Post-v1.0, branch first.
- Body: name the user-visible effect + note it's `go.mod`/`go.sum` only. Keep the message identical
  across apps when it's the same release.

## 3. Report — and stop

Summarize per app: new forge version, build/test/lint status, commit SHA. Then **stop** — list what's
unpushed and offer to push; don't push without an explicit go-ahead.

## Gotchas (learned the hard way)

- cwd resets between shell calls → always `cd ../<tool> && …` in one invocation.
- The tag must be **pushed** before the apps can `go get` it.
- Stage **only** `go.mod`/`go.sum` — never sweep in WIP.
- Run `go`/lint from inside each app dir so mise selects that app's pinned toolchain.

## References

`docs/guides/new-tool.md` (bootstrap a consumer), the **new-tool** skill (keep the tool list in sync),
[ADR-0007](docs/adr/0007-public-api-surface-and-stability.md) (the contract — a breaking forge change
needs a coordinated bump, which is what this skill performs).
