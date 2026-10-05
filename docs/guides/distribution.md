# Distribution & release

Canonical build / publish / install model shared by the `github.com/adaouat/*` CLIs
(`bifrost`, `heraut`, `hermes`, and future tools).

forge is a **library** — it ships no goreleaser config of its own. This guide and the
annotated [`goreleaser.sample.yml`](goreleaser.sample.yml) are the **template** each app
copies into its own `.goreleaser.yml` (replace `<app>` with the binary name). OSS GoReleaser
has no remote-include (`includes:` is Pro-only), so this is a copy-and-adapt convention, not a
live dependency.

## The model

- **GoReleaser v2**, `builds.main: ./cmd/<app>/`.
- **Version injection** via `ldflags: -s -w -X main.Version={{ .Tag }}`. Mandatory: the
  `updatecheck` hint is silent on a `dev` version, so released binaries must carry their tag.
- **Raw versioned binaries** — `archives.formats: [binary]`, no tar/zip wrapper. The asset name
  is `<app>_{{ .Version }}_{{ .Os }}_{{ .Arch }}` (from `archives.name_template`); `builds.binary`
  is **plain `<app>`** so the Homebrew cask installs the binary under that name. Checksums cover
  the binaries directly.
- **A cask-only `homebrew` archive** (`tar.gz`) carrying shell completions and the man page,
  pre-generated in `before.hooks` from `<app> completion` / `<app> man`. It exists only for the
  cask; the raw binary stays the user-facing download.
- **One SBOM per binary** (goreleaser `sboms`, via syft — pin it in the app's mise config).

### Why raw binaries

heraut chose this in its ADR-0013 (*Raw Binary GoReleaser Format*). The original driver —
avoiding ~70 lines of tar/zip extraction and the zip-slip surface **in the self-updater** — is
now historical: the self-updater was removed (forge [ADR-0005](../adr/0005-updates-via-package-managers.md),
M5.2). Raw binaries are retained because they keep the curl install a one-liner, checksum what
users actually execute, and need no extraction step — **mise and Homebrew both consume them
directly**. The `homebrew` archive (ADR-0013's 2026-09-26 note) is the one exception, and only
because a cask's static `completions:`/`manpages:` fields need files to point at — curl and mise
users get the same from `<app> completion` / `<app> man`.

## Release ownership

**heraut owns the GitHub Release** for every tool: goreleaser is build-only
(`release: disable: true`, run with `--skip=publish,announce,validate`), and
`heraut release --set-version "$VERSION"` creates the release, tag and changelog (heraut
ADR-0018, build-then-release). bifrost, heraut and hermes all follow this model.

Each tool keeps its own `release.yml` — that file is the Sigstore signer of its packslip and
attestations — but the steps live in three forge composite actions
([ADR-0009](../adr/0009-release-setup-composite-action.md),
[ADR-0015](../adr/0015-shared-release-build-and-publish.md)). heraut adds its Pkl package and a
version sanity check between them, plus separate Docker/GHCR jobs.

## The release workflow

Order matters. The release heraut creates is **immutable** — assets cannot be added after it is
published (`gh release upload` fails with 422) — so everything that ends up on the release must
be in `dist/` before `heraut release` runs. The three actions encode that order:

| Action | Steps |
|---|---|
| [`release-setup`](../../.github/actions/release-setup/action.yml) | mise, the bootstrap heraut, GPG, bot identity, resolved `$VERSION` (a manual override is normalized: `1.2.3` → `v1.2.3`) |
| [`release-build`](../../.github/actions/release-build/action.yml) | goreleaser build-only (forge owns the exact version) → copy versioned binaries out of `dist/artifacts.json` → `actions/attest` over `checksums.txt` → packslip manifest into `dist/` (both skipped for a non-public repository; see [Private repositories](#private-repositories)) |
| [`release-publish`](../../.github/actions/release-publish/action.yml) | `heraut check` → `heraut release --set-version "$VERSION"` (with `GH_TOKEN` *and* `GITHUB_TOKEN`, which heraut's PR attribution reads) → push the generated cask to the tap (skipped without a token) |

They run as steps of the tool's own job — never wrap them in a reusable workflow: packslip and
the attestation would then be signed by forge's workflow, and mise refuses such a bundle
(ADR-0015). A tool's release job:

```yaml
jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write # release, tag, assets
      id-token: write # Sigstore signing (attest, packslip)
      attestations: write
    env:
      HERAUT_CHECK_UPDATE: false # the job runs heraut — silence its update check
    steps:
      - uses: actions/checkout@<sha> # v7
        with:
          fetch-depth: 0 # changelog needs full history
          token: ${{ secrets.GITHUB_TOKEN }}
      - uses: adaouat/forge/.github/actions/release-setup@<forge-sha> # <forge-tag>
        with:
          gpg-private-key: ${{ secrets.RELEASE_GPG_PRIVATE_KEY }}
          github-token: ${{ secrets.GITHUB_TOKEN }}
          version: ${{ inputs.version }}
      - uses: adaouat/forge/.github/actions/release-build@<forge-sha> # <forge-tag>
        with:
          app: <app>
          github-token: ${{ secrets.GITHUB_TOKEN }}
      - uses: adaouat/forge/.github/actions/release-publish@<forge-sha> # <forge-tag>
        with:
          app: <app>
          github-token: ${{ secrets.GITHUB_TOKEN }}
          homebrew-tap-token: ${{ secrets.HOMEBREW_TAP_TOKEN }}
          regenerate-changelog: ${{ inputs.regenerate_changelog }}
```

with the `version` / `regenerate_changelog` `workflow_dispatch` inputs from forge's own
[`release.yml`](../../.github/workflows/release.yml). The packslip list is derived from
`artifacts.json`, so a tool building fewer platforms (hermes is macOS-only) needs no extra input.

`.config/heraut.yml` `release.assets` must list everything in `dist/` that ships:

```yaml
release:
  assets:
    - "dist/<app>_*_linux_amd64"
    - "dist/<app>_*_linux_arm64"
    - "dist/<app>_*_darwin_amd64"
    - "dist/<app>_*_darwin_arm64"
    - "dist/<app>_*_windows_amd64.exe"
    - "dist/<app>_*_linux_amd64.tar.gz" # homebrew archive — the cask URL 404s without these
    - "dist/<app>_*_linux_arm64.tar.gz"
    - "dist/<app>_*_darwin_amd64.tar.gz"
    - "dist/<app>_*_darwin_arm64.tar.gz"
    - "dist/checksums.txt"
    - "dist/packslip.sigstore.json"
    - "dist/*.sbom.json"
```

## Install channels

All channels consume the same raw binaries.

- **mise** (`packslip` backend — verifies the Sigstore manifest):
  ```bash
  mise use packslip:adaouat/<app>
  ```
  or in `mise.toml`: `"packslip:adaouat/<app>" = "<major>"`. A tool not yet publishing a
  packslip installs with the unverified `github:adaouat/<app>` backend instead.
- **curl**:
  ```bash
  curl -L -o <app> https://github.com/adaouat/<app>/releases/latest/download/<app>_<version>_<os>_<arch>
  chmod +x <app> && sudo mv <app> /usr/local/bin/
  ```
- **Homebrew** (`brew install --cask adaouat/tap/<app>`): a shared `adaouat/homebrew-tap` repo;
  each app publishes a **cask** via `homebrew_casks` (the `brews` *formula* form was removed in
  goreleaser v2.16). **Plain `builds.binary` is what makes the cask install as `<app>`** — a
  versioned `builds.binary` makes the cask install under the long name. Because the release is
  build-only, the cask needs an explicit `url.template` and `ids: [homebrew]`, and a
  post-release step pushes it. Completions and the man page come from the cask's static
  `completions:`/`manpages:` fields — never `generate_completions_from_executable`, which runs
  the unsigned binary and hangs under Gatekeeper. Validate the generated cask with
  `goreleaser release --snapshot --clean` before the first real tag.
  - **Gatekeeper.** The binaries aren't Developer ID-signed/notarized, so running a cask-installed
    binary hangs until `com.apple.quarantine` is removed. The sample carries an opt-in
    `hooks.post.install` that strips it — a deliberate bypass of a macOS check, decided per tool
    (heraut has it) until signing+notarization ships.

## Private repositories

The pipeline above assumes a **public** repository. A private tool can use the same three
actions, with these differences ([ADR-0017](../adr/0017-release-build-private-repositories.md)).

- **No provenance.** Artifact attestations in private repositories need GitHub Enterprise Cloud,
  and packslip always signs through the public Sigstore instance, whose Rekor log would publish
  the private repository's workflow identity. `release-build` therefore skips both itself, with
  no input to set: it runs them only when `github.event.repository.visibility` is `public`, and
  otherwise logs a `::notice::`. The build, the collected binaries (`artifacts` output),
  `checksums.txt` and the SBOMs are produced as usual.
- **Permissions.** The release job needs only `contents: write`. Drop `id-token: write` and
  `attestations: write`. Add `pull-requests: read`, because a job-level `permissions:` block
  sets every unlisted scope to none, and on a private repository heraut then cannot read PRs
  for changelog attribution. Its default `optional` enrichment policy degrades to "remote
  metadata unavailable — PR authors/numbers omitted" instead of failing.
- **`release.assets`.** Omit `dist/packslip.sigstore.json` (it is never produced). If it stays,
  heraut's GitHub driver prints `warning: no files matched asset pattern
  "dist/packslip.sigstore.json" — skipping` to stderr and creates the release with the other
  assets. A pattern matching nothing never fails the release; only invalid glob syntax does.
  Also omit the `homebrew` `.tar.gz` archives (see below).
- **No Homebrew cask.** The tap is public and a cask cannot download private release assets
  anonymously. Leave `homebrew_casks` and the `homebrew` archive out of `.goreleaser.yml` and
  call `release-publish` without `homebrew-tap-token` (the push is then skipped).
- **Install through mise's `github:` backend**, not `packslip:` (there is no manifest to verify):
  ```bash
  mise use github:adaouat/<app>
  ```
  It authenticates with `GITHUB_TOKEN` or `MISE_GITHUB_TOKEN`, or with the `gh` CLI login, and
  verifies no provenance (there is none). The curl one-liner does not work either, because
  private release assets need authentication. Use
  `gh release download --repo adaouat/<app> --pattern '<app>_*_<os>_<arch>'` instead.
- **No update hint.** `updatecheck` queries the GitHub API unauthenticated; for a private
  repository that lookup fails, and the `Hinter` stays silent by design.

## Status

- **Homebrew tap** — `adaouat/homebrew-tap` (one cask per tool, generated on release).
- **Lint/test CI** — shared via forge's reusable `go-ci.yml`
  ([ADR-0006](../adr/0006-shared-ci-reusable-workflow.md)).
- **Release workflows** — per-app files calling the three actions (see *Release ownership*).
  Apps move onto `release-build`/`release-publish` once forge is tagged with them (roadmap M18);
  until then each carries the same steps inline.
