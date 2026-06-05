# ADR-0009 — Shared release setup via a composite action

**Status:** Accepted
**Date:** 2026-06-05

## Context

All three repos release via heraut, and their release workflows share an identical six-step
prelude: **checkout → mise → install heraut → import GPG → configure git identity → resolve
version**. After that they diverge — forge stops (library); bifrost adds goreleaser build +
collect + attest + Homebrew cask; heraut adds those plus a version sanity check, Docker jobs, and
uses its self-built binary for the release. So the *setup* is shared (3 consumers, identical); the
*build/publish* is not.

A reusable `workflow_call` workflow (like [go-ci.yml](0006-shared-ci-reusable-workflow.md)) is the
wrong tool here: it runs as a **separate job**, so the caller's build/release steps — which must
run in the same job, on the same checkout, with the same GPG/git state and the resolved-version
env — couldn't follow it.

## Decision

Extract the shared prelude as a **composite action**, `forge/.github/actions/release-setup`,
consumed by all three repos (forge via the local `./…` path; the apps via
`adaouat/forge/.github/actions/release-setup@<sha>`).

- It runs as **steps inside the caller's job**, so each app's build/release/cask/Docker steps
  follow it naturally, sharing the workspace and state.
- **Checkout stays in the calling workflow.** A local-action reference (`./…`, which forge itself
  uses) needs the repo checked out before the action is resolvable; the apps' cross-repo reference
  fetches the action automatically. So the action covers the five steps *after* checkout.
- **Inputs:** `gpg-private-key`, `github-token`, `version` (override). **Outputs:** `version`,
  `tag`. It also exports `VERSION`/`VERSION_OVERRIDDEN` to `GITHUB_ENV`, so existing caller steps
  keep using `$VERSION` unchanged.
- Build/publish (goreleaser, collect, attest, cask, Docker) stays **per-app** — genuinely
  divergent, so not forced into a toggle-laden shared workflow (forge's YAGNI bar).

## Consequences

- One home for the identical setup, and for the mise / heraut-download / gpg-import action SHA
  pins (bump once, not in three files — checkout's pin stays per-workflow since it lives outside).
- A third tool inherits the whole setup with a single `uses:` step.
- **Cross-repo coupling:** an app's release now depends on forge's action at the pinned SHA — the
  same trade-off as the go-ci.yml reusable workflow (ADR-0006).
- The boundary is the clean 3-of-3-identical part; preflight + `heraut release` stay per-app
  (heraut uses its self-built binary, not the downloaded one) — they could fold into an optional
  second step later if it earns its keep.
