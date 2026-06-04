# Distribution & release

Canonical build / publish / install model shared by the `github.com/adaouat/*` CLIs
(`bifrost`, `heraut`, and future tools).

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

### Why raw binaries

heraut chose this in its ADR-0013 (*Raw Binary GoReleaser Format*). The original driver —
avoiding ~70 lines of tar/zip extraction and the zip-slip surface **in the self-updater** — is
now historical: the self-updater was removed (forge [ADR-0005](../adr/0005-updates-via-package-managers.md),
M5.2). Raw binaries are retained because they keep the curl install a one-liner, checksum what
users actually execute, and need no extraction step — **mise and Homebrew both consume them
directly**. Switching to archives would still work with mise (its `github`/ubi backend
auto-extracts `.tar.gz`/`.zip`), but adds a `tar xz` step to curl for no benefit, since nothing
is bundled alongside the binary (completions come from a subcommand).

## Release ownership

heraut is the family's release tool, so the canonical end state is **heraut owns the GitHub
Release**:

- **heraut-owned** — goreleaser is build-only (`release: disable: true`); heraut creates the
  release (`heraut release --version`) and additionally publishes a Docker image to GHCR.
- **Self-release (interim)** — goreleaser cuts the release itself (`release: disable: false` /
  omitted). An app stays here until heraut-driven release is wired for it. bifrost is here today.

Release *workflows* are **not** shared: heraut's (self-release + Docker) and bifrost's
(self-release only) are too divergent to unify now.

## Install channels

All channels consume the same raw binaries.

- **mise** (`github` backend — the former `ubi`):
  ```bash
  mise use github:adaouat/<app>
  ```
  or in `mise.toml`: `"github:adaouat/<app>" = "latest"`. The backend matches the os/arch
  tokens in the asset name and installs the binary directly.
- **curl**:
  ```bash
  curl -L -o <app> https://github.com/adaouat/<app>/releases/latest/download/<app>_<version>_<os>_<arch>
  chmod +x <app> && sudo mv <app> /usr/local/bin/
  ```
- **Homebrew** (`brew install --cask adaouat/tap/<app>`): a shared `adaouat/homebrew-tap` repo;
  each app publishes a **cask** via `homebrew_casks` (the `brews` *formula* form is deprecated for
  pre-built binaries). goreleaser generates a per-platform cask (`on_macos` / `on_arm`, `binary
  …, target: "<app>"`). **Plain `builds.binary` is what makes the cask install as `<app>`** — a
  versioned `builds.binary` makes the cask install under the long name. Always validate the
  generated cask with `goreleaser release --snapshot --clean` before the first real tag. Two
  cases, by release ownership:
  - **goreleaser-owned release** (bifrost): the default download URL works. The block is just
    `repository` + `directory: Casks` + `homepage` + `description` + `token` (see the sample),
    and goreleaser pushes the cask during the release.
  - **build-only release** (heraut, `release: disable: true`): goreleaser can't derive the URL, so
    set an explicit `url.template` pointing at the release assets. goreleaser only *generates* the
    cask (it runs `--skip=publish`), so a **post-release workflow step pushes it** to the tap after
    the assets are uploaded (skip it gracefully when the token is unset). Plain `builds.binary`
    also means build outputs aren't versioned on disk — map them to the versioned asset names via
    goreleaser's `artifacts.json` in the collect step.

## Status

- **Homebrew tap** — `adaouat/homebrew-tap` (one cask per tool, generated on release).
- **Lint/test CI** — shared via forge's reusable `go-ci.yml`
  ([ADR-0006](../adr/0006-shared-ci-reusable-workflow.md)).
- **Release workflows** — stay per-app (too divergent: heraut self-release + Docker/GHCR,
  bifrost goreleaser-owned).
