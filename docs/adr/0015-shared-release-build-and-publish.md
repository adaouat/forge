# ADR-0015 — Shared release build and publish via composite actions

**Status:** Accepted
**Date:** 2026-10-04

## Context

[ADR-0009](0009-release-setup-composite-action.md) shared only the release *prelude*
(`release-setup`) and kept build/publish per-app as "genuinely divergent". That is no longer
true. Since M18 every tool releases the same way — goreleaser build-only, heraut owns the
release (heraut ADR-0018) — and the steps after `release-setup` are the same in all three:

**build** (goreleaser, exact version pin) → **collect** (copy versioned binaries out of
`dist/artifacts.json`) → **attest** (`actions/attest` over `checksums.txt`) → **packslip** (sign
the raw binaries, `upload: false`) → **preflight** (`heraut check`) → **release**
(`heraut release --set-version`) → **cask push** to `adaouat/homebrew-tap`.

bifrost's and hermes's `release.yml` are byte-identical apart from the tool name and hermes's
macOS-only packslip artifact list. heraut runs the same steps with three additions placed
*between* them: a Pkl package (after build), a version sanity check (before preflight), and
calling its freshly built binary instead of the downloaded one; its Docker jobs are separate jobs.

The cost of copies showed up in M18. heraut picked up packslip, SBOMs, the `GITHUB_TOKEN` fix
and the goreleaser pin; the other two received none of them until they were ported by hand, and
a heraut flag rename (`--version` → `--set-version`) broke all three release workflows unnoticed.

### Options

1. **Status quo — per-app copies** from a documented template (`distribution.md`). No coupling,
   but this is the drift M18 just paid for, and it grows with each new tool.
2. **A reusable `workflow_call` release workflow** (as [ADR-0006](0006-shared-ci-reusable-workflow.md)
   did for `go-ci.yml`). **Not possible for this pipeline.** A reusable workflow's OIDC identity is
   *its own* file: packslip run inside a reusable workflow hosted in another repository signs as
   `adaouat/forge/.github/workflows/…`, so packslip's own verification fails and consumers (mise)
   refuse the bundle — they accept only the project's own workflows (packslip *Publishing*
   docs; spec "Reusable workflows"). Build provenance attestations would likewise name forge's
   workflow. It would also split the pipeline across jobs, which ADR-0009 already ruled out.
3. **A canonical template file** (`docs/guides/release.sample.yml`) copied by each tool. Makes
   new tools easy, but existing tools still drift — the same failure as option 1.
4. **More composite actions**, following ADR-0009. Composite actions run *as steps of the
   caller's job*, so the OIDC token (and therefore the packslip and attestation identity) stays
   the app's own `release.yml`, and the shared workspace/GPG/`$VERSION` state carries over.

## Decision

Option 4. Add two composite actions next to `release-setup`, split where heraut
inserts its own steps:

- **`release-build`** — goreleaser build (forge owns the exact goreleaser version) → collect →
  attest → packslip. Inputs: `app` (binary name). The packslip artifact list is derived from
  `dist/artifacts.json` (every `Binary` artifact), so a macOS-only tool such as hermes needs no
  special-casing. Requires the caller's job to grant `contents: write`, `id-token: write` and
  `attestations: write`.
- **`release-publish`** — preflight → `heraut release --set-version "$VERSION"` (with both
  `GH_TOKEN` and `GITHUB_TOKEN`) → cask push. Inputs: `app`, `github-token`,
  `homebrew-tap-token` (optional, skipped when empty, as today), `regenerate-changelog`, and
  `heraut-bin` (default `heraut`; heraut passes its fresh build).

A tool's release job becomes: checkout → `release-setup` → `release-build` → *(tool-specific
steps)* → `release-publish`. heraut keeps its Pkl and sanity-check steps between the two, and its
Docker jobs unchanged. Each tool keeps its own `release.yml` — which is also what keeps the
signing identity per-repo.

## Consequences

- One place to bump the goreleaser, packslip and attest pins (checkout stays per-workflow), and to
  change the pipeline (a new asset, a heraut flag change) — the M18 class of drift goes away for
  all three repos.
- **Three real consumers, identical behavior** — clears the [ADR-0001](0001-shared-core-module.md)
  bar. The parameters are a tool name, a token and a binary path, not behavior toggles.
- **Coupling grows** in the same way as ADR-0006/0009: an app's release depends on forge's actions
  at the pinned SHA. Mitigated by SHA pinning — apps upgrade deliberately. forge's own release
  calls `release-publish` (no cask token, so the push is skipped), which exercises preflight and
  `heraut release` on every forge release; `release-build` is only exercised by the tools, so a
  tool's first release after re-pinning is its real test.
- **Signing identity is unchanged**: each artifact is still signed by `adaouat/<tool>`'s
  `release.yml`, so existing mise `packslip:` pins and lockfile signers keep verifying. Renaming a
  tool's `release.yml` would still be a signer change consumers must approve — unrelated to this
  ADR, but worth knowing.
- `release.assets` in each tool's `.config/heraut.yml` stays per-tool (it is heraut config, not CI).
- `distribution.md` and the `new-tool` skill shrink to "call the three actions"; the step-by-step
  section becomes the actions' own documentation.
