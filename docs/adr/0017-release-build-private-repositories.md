# ADR-0017 — `release-build` skips provenance in private repositories

**Status:** Accepted (amends [ADR-0015](0015-shared-release-build-and-publish.md))
**Date:** 2026-10-05

## Context

[ADR-0015](0015-shared-release-build-and-publish.md) moved build → collect → attest → packslip
into the `release-build` composite action. Every caller so far is a public repository. A
**private** tool on a plan without GitHub Enterprise Cloud cannot release through it, which
the sources (read at the pinned SHAs) confirm:

- **`actions/attest` fails.** Its README: *"To use artifact attestations in private or internal
  repositories, you must be on a GitHub Enterprise Cloud plan."* For any repository whose
  `payload.repository.visibility` is not `public`, it signs through GitHub's private Sigstore
  instance (`src/main.ts`), which is that Enterprise Cloud feature.
- **packslip's attest sub-step fails the same way.** The `jdx/packslip` action's own `attest`
  input defaults to `true`, which runs `actions/attest-build-provenance` over each binary. That
  is why a public release today carries two attestations: one over `checksums.txt` and one per
  binary.
- **packslip's manifest signing would succeed, but leak.** packslip always signs with the
  *public* Sigstore instance (`SigningContext::production()`, `SIGSTORE_PRODUCTION_TRUSTED_ROOT`)
  and logs to the public Rekor transparency log. On a private repository that permanently
  publishes a certificate naming the repository's release workflow, ref and commit, so turning
  off only its attest sub-step would trade a failure for a disclosure.

The visibility signal is reliable where it matters. `actions/attest` itself keys on
`github.event.repository.visibility`. heraut's `workflow_dispatch` release of 2026-10-04 logged
*"signed using certificate from Public Good Sigstore instance"*, which `actions/attest` only
chooses when that field reads `public`, so the field is present under `workflow_dispatch`.

## Decision

`release-build` gates both provenance steps on the caller's visibility, with **no new input**:

- `Attest build provenance` and `Publish packslip` run only when
  `github.event.repository.visibility == 'public'`. This is the same rule `actions/attest` uses,
  so `private` and `internal` repositories are both treated as not public, and so is a missing
  field (fail safe).
- Otherwise a step emits a `::notice::` that provenance was skipped and why.
- Unchanged in every case: the goreleaser build, the *Collect release binaries* step (still
  copies the versioned binaries into `dist/` and still outputs `artifacts`), `checksums.txt` and
  the SBOMs.

### Why no `attest` input

An explicit `auto | true | false` input was considered and rejected. ADR-0015's parameters are
*"a tool name, a token and a binary path, not behavior toggles"*, and neither forced mode has a
real use. A forced `true` on a private repository fails without Enterprise Cloud, and even with
it packslip would still publish the repository's identity to the public log. A forced `false`
on a public repository would only drop the provenance that the mise `packslip:` install channel
depends on. The skip follows a platform constraint, not a per-tool preference, so it is derived,
not configured. If Enterprise Cloud ever appears, attestation alone (not packslip) could be
re-enabled for private repositories; that would be a new ADR.

## Consequences

- **Public callers are unchanged.** forge, bifrost, heraut and hermes see the same steps run in
  the same order with the same pins; only the skip notice step is added, and it does not run
  for them.
- **A private tool ships no packslip manifest and no attestation**, so the mise `packslip:`
  backend cannot install it. It installs through mise's `github:` backend, authenticated by a
  token, which verifies no provenance. It also gets no Homebrew cask (the tap is public and a
  cask cannot download private assets anonymously).
- **The private caller's job needs only `contents: write`**, not `id-token` or
  `attestations`. Its `.config/heraut.yml` should drop `dist/packslip.sigstore.json` from
  `release.assets`. Leaving it is harmless, since heraut warns and skips a pattern that
  matches nothing.
- Nothing in `release-setup` or `release-publish` needs a change for a private caller. The
  bootstrap `gh release download --repo adaouat/heraut` reads a public repository, and the cask
  push is already skipped without a tap token.
- As with ADR-0015, `release-build` is only exercised by a tool's release, so the first release
  of a private tool is the real test of the skip path.
- The duplicate attestation on public releases (forge's over `checksums.txt`, packslip's per
  binary) is noted but not changed here. Removing either would change public release output.
