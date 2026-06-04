# ADR-0005 — Updates via package managers, not binary self-replacement

**Status:** Accepted
**Date:** 2026-06-04

## Context

[ADR-0001](0001-shared-core-module.md) and roadmap M5 planned to extract heraut's hand-rolled
self-updater — GitHub Releases API fetch + SHA-256 verify + atomic `os.Rename` replace (with a
`.old` backup and a sudo hint) + a daily "newer version" hint — generalised over repo and asset
naming.

Two things argue against carrying that forward:

1. **Self-replacement is the bug-prone, security-sensitive part.** Atomically swapping a
   running executable is platform-specific (permissions, Windows file locking, rollback, TLS,
   checksum). heraut's implementation accrued real bugs that had to be found and fixed.
2. **The Go self-update library landscape is thin.** Measured dependency trees:
   - `creativeprojects/go-selfupdate` — **51 modules**; pulls `google/go-github/v74` *and*
     GitLab *and* Gitea SDKs even when only GitHub is used. Maintained, but heavy for a
     foundation library imported by every CLI.
   - `minio/selfupdate` — **9 modules**, maintained, but carries maintainer-trust risk (MinIO's
     history of relicensing and gutting open-source features).
   - `inconshreveable/go-update` (the original `minio/selfupdate` forked) — **1 module, zero
     deps**, but frozen at a 2016 commit.

   There is no lightweight + maintained + trusted option.

Meanwhile the project already standardises on **mise** and **goreleaser**, which install and
upgrade GitHub-release binaries with zero custom code.

## Decision

**Do not self-replace the binary.** Distribution and upgrades are delegated to package
managers; forge owns only the safe "is there a newer version, and how do I get it" half.

- **Upgrades** go through whatever installed the binary: goreleaser publishes GitHub releases +
  a Homebrew tap + Scoop manifests; mise installs/upgrades via its **`github`** backend
  (`mise use github:adaouat/<app>`; `mise upgrade <app>`); a curl install script covers the
  no-package-manager case (re-running it overwrites).
- **forge ships an `updatecheck` package** (renamed from the roadmap's `selfupdate`, since it
  performs no update) that does only:
  - `CheckLatest(ctx, repo, current)` — `GET …/releases/latest` + semver compare. **`net/http`
    + `encoding/json` only; it never downloads or touches the binary.**
  - **Install-method detection** from the running executable's real path (resolve symlinks):
    Homebrew (`/Cellar/`), mise (`/mise/installs/`), Scoop (`\scoop\…`), `go install`
    (`$GOBIN`/`$GOPATH/bin`), with a generic fallback. Parameterised over binary name and
    module path.
  - A 24h-cached `Hint` that prints `"<app> X.Y.Z available — run: <detected upgrade command>"`,
    swallowing all errors so it never breaks a normal run.

This **supersedes the `selfupdate` Tier-1 scope in ADR-0001** (no atomic replace / SHA-256 /
asset download).

## Consequences

- **Zero binary-replacement code and bug surface**, and **no risky, heavy, or untrusted
  dependency** — the package is stdlib-only (`net/http` + `encoding/json` + `os`/`filepath`).
- forge owns the version check + install detection; both CLIs and the incoming third tool share
  it, and **bifrost gains the update hint for free** (it had no self-update before).
- **heraut's `internal/selfupdate` binary replacer is removed.** Its `self-update` command
  becomes informational (prints the detected upgrade command) or is dropped — the daily hint now
  comes from forge.
- **New release-side responsibility (app-side, not forge runtime):** the apps' goreleaser config
  must publish the Homebrew tap + Scoop manifests, and the install docs must cover the mise
  `github` backend and the curl script. forge stays a pure library.
- Install detection is a **heuristic with a generic fallback** — it proposes a tailored command
  when confident, otherwise points at the releases page / install script. `apt`/`rpm` are not
  detected in v1 (would require shelling out to `dpkg -S`/`rpm -qf`).
