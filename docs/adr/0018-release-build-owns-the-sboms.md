# ADR-0018 — `release-build` generates CycloneDX SBOMs to BSI TR-03183-2

**Status:** Proposed (amends [ADR-0015](0015-shared-release-build-and-publish.md))
**Date:** 2026-11-02

## Context

Every tool ships one SBOM per raw binary (`<asset>.sbom.json`). Each app's `.goreleaser.yml`
makes them with a `sboms:` block that runs syft at its default output, **SPDX 2.3**. Each app also
pins syft in its own mise config (heraut `latest`, the others `1.54`). That is the per-app
drift ADR-0015 set out to remove.

heraut v0.82.1's darwin/arm64 SBOM, scored with sbomqs (the engine behind demo.interlynk.io)
against **BSI TR-03183-2 v2.1**, gets 3.3–5.0/10, grade F. BSI's guideline is the strictest
public SBOM profile and the usual EU CRA yardstick. The gaps:

- **Format.** BSI v2.1 only accepts CycloneDX ≥ 1.6 or SPDX ≥ 3.0.1. SPDX 2.3 fails outright.
  syft writes CycloneDX 1.6 (`-o cyclonedx-json@1.6`), not SPDX 3.
- **No identity.** The document and every component have the supplier `NOASSERTION`, with no
  email or URL.
- **Wrong primary component.** The top-level entry is the binary file, versioned by its sha256,
  with no licence. The main module's version is a Go pseudo-version
  (`v0.82.1-0.20261101161728-20ed79dd17e7`). Go stamps it because the build runs before
  `heraut release` creates the tag (heraut ADR-0018), and `GORELEASER_CURRENT_TAG` does not
  change what Go stamps. The binary's `--version` is correct (ldflags); only the SBOM is wrong.
- **Fields no generator writes:**
  - the BSI fields for a component's filename and its executable, archive and structured flags
    (`bsi:component:*` properties);
  - a SHA-512 hash of each component's deployable form;
  - licences marked `acknowledgement: concluded`;
  - source code and download URLs.

Switching syft to CycloneDX alone scores *lower* (2.8), because its CycloneDX output also drops
the licences it found. A post-processing prototype, written in jq over syft's CycloneDX 1.6
output, scored **9.4/10 (A)** on the same binary, using only values that can be checked
(`docs/plans/m23-sbom/enrich.jq`, run over `docs/plans/m23-sbom/syft-cyclonedx.json`).

The CRA itself does not mandate BSI TR-03183-2. It asks for a machine-readable SBOM in a common
format, covering at least the top-level dependencies, and its main obligations apply from
11 December 2027. BSI v2.1 is the target because it is the strictest; meeting it meets the
lighter profiles (NTIA, CISA) too.

**Constraint: the SBOMs are attested.** goreleaser lists them in `checksums.txt`, and
`release-build` attests `checksums.txt` (ADR-0015). Changing an SBOM after goreleaser must
therefore update its line in `checksums.txt` before the *Attest build provenance* step.

## Decision

`release-build` generates, enriches and checksums the SBOMs itself, in one new step after
*Collect release binaries* and before *Attest build provenance*. It takes **no new input**, in
line with ADR-0015's "no behaviour toggles"; everything is derived from the run.

1. **Generate.** For each collected binary `dist/<asset>`, run syft, pinned by forge (the
   action installs the exact version, as it does goreleaser), with
   `-o cyclonedx-json@1.6=dist/<asset>.sbom.json`. Licence search in the local Go module cache
   is turned on (`SYFT_GOLANG_SEARCH_LOCAL_MOD_CACHE_LICENSES`), since goreleaser's build has
   just filled that cache.
2. **Enrich** with a jq filter shipped next to the action:
   - **Document:**
     - The author, manufacturer and supplier are the repository owner, with
       `https://github.com/<owner>` as their URL.
   - **Primary component.** It is the main Go module, renamed `<app>`, type `application`:
     - Its version is `$VERSION`, which replaces the pseudo-version, and its purl is
       `pkg:golang/<module>@$VERSION`.
     - Its licence is the repository's SPDX id from the GitHub API, listed twice, once as
       `declared` and once as `concluded`.
     - Its external references are the repository as `vcs`; the release download URL of
       `<asset>` as `distribution`, with the binary's SHA-512; and
       `https://github.com/<repo>/security` as `security-contact`.
     - Its BSI properties are filename `<asset>`, `executable`, `no archive`, `unstructured`.
     - syft's separate file component goes. The dependency graph is re-rooted on the primary
       component.
   - **Each Go module dependency:**
     - **Supplier:** named after the first two segments of the module path, with
       `https://<module path without /vN>` as its URL.
     - **Licences:** syft's licences, each listed once as `declared` and once as `concluded`.
       syft found them by scanning the module's licence file.
     - **External references:** `vcs` is that same URL. `distribution` is
       `https://proxy.golang.org/<module>/@v/<version>.zip`, with the SHA-512 of that zip from
       the module cache, the zip Go has already checked against `go.sum`.
     - **BSI properties:** filename `<version>.zip`, `non-executable`, `archive`, `structured`.
       They describe the module zip, which is the form a Go dependency is distributed in.
     - A module with no zip in the cache keeps its other fields and gets no hash. The step
       logs a `::notice::` for it.
3. **Checksum.** Write each SBOM's `sha256  <name>` line into `dist/checksums.txt`: replace
   the line if goreleaser already wrote one, append it otherwise. The attestation then covers
   the enriched files.

The file name stays `<asset>.sbom.json`, so every app's `release.assets` glob
`dist/*.sbom.json` keeps matching. The apps then drop their own `sboms:` block and syft pin.
Until an app has done so, its goreleaser SBOMs are overwritten (same name) or left next to
forge's. Windows is the exception: goreleaser named that one `heraut.exe_<v>_windows_amd64`,
forge names it `<asset>.sbom.json`. Either way the release is not broken, and step 3 keeps
`checksums.txt` right.

### Rejected

- **Leave generation in each app and only switch the syft output format.** It scores lower than
  today, and it keeps four syft pins and the mise-shim workaround (`MISE_TRUSTED_CONFIG_PATHS`
  in `sboms[].env`).
- **Run the enrichment as goreleaser's `sboms[].cmd`, pointing at a script in forge.**
  goreleaser runs that command with a minimal environment, so the script's path and
  `$VERSION` would have to be wired through each app's config. A local
  `goreleaser --snapshot` run would also break.
- **sbomasm (Interlynk) for the enrichment.** Its `edit` command covers the document and the
  primary component, but not the per-dependency BSI fields, hashes or licence acknowledgement.
  It would be a second pinned tool on top of jq, and jq is already on the runners.
- **A minimum sbomqs score as a release gate.** That would tie releases to a third-party scorer
  whose checks change between versions (the demo site and the latest sbomqs already disagree
  on one check). The score is checked when the filter changes, not on every release.
- **SPDX 3.0.1.** syft cannot write it.

## Consequences

- One syft version, in forge, for every tool. The per-app `sboms:` block, the syft mise pins
  and the mise-shim workaround go away.
- The SBOM finally names the released version and the supplier, which also helps any scanner
  (Grype, osv-scanner) that matches on purl.
- The release format changes from SPDX 2.3 to CycloneDX 1.6. Anyone parsing the old format
  must adapt. Every release so far has been in SPDX, and no consumer is known.
- **Supplier approximation.** A dependency's supplier is derived from its module path. For
  vanity import paths (`charm.land/...`, `golang.org/x/...`) the URL is the vanity host, which
  is real but is not the owner's own page.
- **Private repositories.** A private tool's modules served outside proxy.golang.org (`GOPRIVATE`)
  would get proxy URLs that don't resolve. No current tool has a private dependency. If one
  does, its modules' `distribution` reference is dropped then, not designed for now.
- `release-build` is only run by a release, so the first release after adoption is the real
  test. The filter itself is tested before that, against a recorded syft SBOM: a fixture test
  plus an sbomqs `bsi-v2.1` score of at least 9.0, run when the filter changes.
- Not legal advice. Nothing here makes a tool "CRA-compliant"; the SBOM requirement is one
  item among the regulation's essential requirements.
