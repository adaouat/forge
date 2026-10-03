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

Release *workflows* stay per-app. bifrost's and hermes's are identical modulo the tool name;
heraut's adds a Docker/GHCR image and its Pkl package. Whether to share the common part is an
open roadmap decision (M18).

## The release workflow

Order matters. The release heraut creates is **immutable** — assets cannot be added after it is
published (`gh release upload` fails with 422) — so everything that ends up on the release must
be in `dist/` before the `Release` step.

1. **Checkout** with `fetch-depth: 0` (changelog needs full history).
2. **Release setup** — forge's [`release-setup`](../../.github/actions/release-setup/action.yml)
   composite ([ADR-0009](../adr/0009-release-setup-composite-action.md)): mise, the bootstrap
   heraut, GPG, bot identity, and the resolved `VERSION`. A manual version override is
   normalized there (`1.2.3` → `v1.2.3`) so goreleaser never bakes an unprefixed tag.
3. **Build** — `goreleaser/goreleaser-action` with an **exact** `version:` pin and
   `GORELEASER_CURRENT_TAG: ${{ env.VERSION }}` (the tag doesn't exist yet). Avoid goreleaser
   2.18.0–2.18.1's not-yet-tagged-tag regression (goreleaser#7124, fixed in 2.18.2).
4. **Collect** — `builds.binary` is plain, so copy each build output to its versioned asset name
   using `dist/artifacts.json`.
5. **Attest** — `actions/attest` over `dist/checksums.txt` (needs `id-token: write` +
   `attestations: write`).
6. **packslip** — `jdx/packslip` signs a Sigstore manifest over the raw binaries so a
   packslip-aware installer (mise) verifies what it downloads:
   ```yaml
   - name: Publish packslip
     uses: jdx/packslip@87479dfc6443253dff69601cace5fc6ea07e6df5 # v1.4.0
     with:
       artifacts: >-
         dist/<app>_*_linux_amd64 dist/<app>_*_linux_arm64 dist/<app>_*_darwin_amd64 dist/<app>_*_darwin_arm64 dist/<app>_*_windows_amd64.exe
       bin: <app>
       tag: ${{ env.VERSION }} # workflow_dispatch: the action can't infer the tag
       out: dist
       upload: false # heraut uploads it with the other assets (immutable release)
   ```
   Use v1.4.0 or later: earlier versions record `CGO_ENABLED=0` Linux builds as glibc-only, so
   mise on musl (Alpine) finds no matching asset.
7. **Preflight** (`heraut check`) and **Release** (`heraut release --set-version "$VERSION"`).
   Both need `GITHUB_TOKEN` *as well as* `GH_TOKEN` — heraut's PR attribution reads
   `GITHUB_TOKEN`; without it those GraphQL calls go unauthenticated and hit the 60/hr limit.
8. **Push the cask** to the tap (goreleaser only generated it). Skip gracefully when
   `HOMEBREW_TAP_TOKEN` is unset.

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

## Status

- **Homebrew tap** — `adaouat/homebrew-tap` (one cask per tool, generated on release).
- **Lint/test CI** — shared via forge's reusable `go-ci.yml`
  ([ADR-0006](../adr/0006-shared-ci-reusable-workflow.md)).
- **Release workflows** — per-app (see *Release ownership*). heraut has every step above;
  bifrost and hermes don't yet have packslip, SBOMs, the `homebrew` archive or `GITHUB_TOKEN`.
