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
- **Raw versioned binaries** — `archives.formats: [binary]`, no tar/zip wrapper. Binary name
  `<app>_{{ .Version }}_{{ .Os }}_{{ .Arch }}`; checksums cover the binaries directly.

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
- **Homebrew** *(tap not yet created)*: a shared `adaouat/homebrew-tap` repo; each app's
  `brews:` block targets it. goreleaser generates a per-platform formula from the raw binaries
  (`on_macos` / `on_arm` / `on_intel`, `bin.install … => "<app>"`). Validate the generated
  formula with `goreleaser release --snapshot --clean` before the first real tag — raw-binary
  formulae are fussier than tarball ones. Skeleton to add to `.goreleaser.yml` once the tap
  exists:
  ```yaml
  brews:
    - repository:
        owner: adaouat
        name: homebrew-tap
      directory: Formula
      homepage: "https://github.com/adaouat/<app>"
      description: "<one-line description>"
      # Finalize install/test stanzas against the --snapshot formula output.
  ```

## Not shared (yet)

- **Homebrew tap repo** — planned (`adaouat/homebrew-tap`), not yet created.
- **Lint/CI reusable workflow** — wanted, tracked on the roadmap as a separate track; it is
  CI plumbing, not part of the release model.
