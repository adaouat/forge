# Starting a new tool on forge

forge is the foundation every `adaouat/*` CLI is built on — it brings the CLI framework (cobra +
fang), the family theme, and the shared runtime (exec, config, exit codes, update check). A new
tool imports forge and wires a thin `main`; forge does the heavy lifting.

## What forge gives you

| Package | What |
|---|---|
| `cli` | `Run(ctx, cmd, version, accent)` — runs your cobra command through fang with the version + theme |
| `ui` | the theme (`Accent`, `ColorScheme`, `HuhTheme`, `DefaultAccent`), status helpers, `Spinner`, header/version renderers, output `Mode`, color/TTY detection, the shared `Palette` |
| `exec` | `Runner` (`Run`/`RunEnv`/`RunDir`) + `CmdRunner`; `exec/exectest` mocks for tests |
| `exitcode` | the shared exit-code vocabulary + `Resolve` / `Wrap` |
| `config` | strict YAML loader, app-name path resolver, `ValidationError` |
| `updatecheck` | the daily "newer release" hint + install-method detection |

Contract: [ADR-0007](../adr/0007-public-api-surface-and-stability.md). forge as the CLI framework
foundation: [ADR-0010](../adr/0010-cli-framework-foundation.md).

## 1. Scaffolding first (Tier-2)

Set up the family's tooling **before** writing any Go — `.config/mise` is what installs the
**pinned Go version** (and golangci-lint, hk, …), so it has to come first. Copy forge's canonical
`.config/{mise,hk,cocogitto,typos,yamlfmt}` and `.claude/rules` (see [tier2-sync.md](tier2-sync.md)):

```bash
mkdir <tool> && cd <tool>
# copy ../forge/.config and adapt ../forge/docs/rules -> .claude/rules
mise trust && mise install   # installs the pinned Go, golangci-lint, hk, …
```

Now `go`, `golangci-lint`, etc. are the exact versions the rest of the family uses.

## 2. Bootstrap

```bash
go mod init github.com/adaouat/<tool>
go get github.com/adaouat/forge@latest
```

`cmd/<tool>/main.go`:

```go
package main

import (
	"context"
	"os"

	"github.com/adaouat/forge/cli"
	forgeexit "github.com/adaouat/forge/exitcode"
	"github.com/adaouat/<tool>/internal/ui"
)

var Version = "dev" // -ldflags "-X main.Version={{.Tag}}"

func main() {
	err := cli.Run(context.Background(), rootCmd(Version), Version, ui.Accent())
	os.Exit(forgeexit.Resolve(err))
}
```

`rootCmd` builds your `cobra.Command` tree as usual. **fang is reached only through `cli.Run`** —
don't import it directly, so its version stays forge's. (cobra you do import, to build commands;
keep its version on forge's pin.)

## 3. Pick an accent

Your tool's one piece of brand. Define it once in `internal/ui` (or `internal/tui`):

```go
import (
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	forgeui "github.com/adaouat/forge/ui"
)

func Accent() forgeui.Accent {
	return forgeui.Accent{
		Light: lipgloss.Color("#…"), Dark: lipgloss.Color("#…"),                     // titles/program/flags
		SecondaryLight: lipgloss.Color("#…"), SecondaryDark: lipgloss.Color("#…"),   // subcommands
	}
}

// HuhTheme themes interactive prompts to match. Use it via form.WithTheme(ui.HuhTheme()).
func HuhTheme() huh.ThemeFunc { return forgeui.HuhTheme(Accent()) }
```

Taken: bifrost **teal/violet**, heraut **gold/azure**, forge's default **Ember** (orange/coal-red).
Pick a distinct hue — or pass `forgeui.Accent{}` to inherit the Ember default. The shared structure
(args blue, muted gray, errors red) comes from the palette either way, so your tool still reads as
family.

## 4. CI

Call forge's shared lint/test workflow ([ADR-0006](../adr/0006-shared-ci-reusable-workflow.md)) in
`.github/workflows/ci.yml`:

```yaml
jobs:
  ci:
    uses: adaouat/forge/.github/workflows/go-ci.yml@<forge-sha> # vX.Y.Z
    with:
      coverage-threshold: <your %>   # required — no default
  build:
    runs-on: ubuntu-latest
    # your go build (+ goreleaser check) — stays per-app
```

## 5. Release + distribution

heraut releases the whole family. Add a `.config/heraut.yml` (versioning + changelog + the github
platform) and a `release.yml` that calls forge's **release-setup** composite action
([ADR-0009](../adr/0009-release-setup-composite-action.md)) and then `heraut release`. The
build/publish model — goreleaser raw binaries, mise / curl / Homebrew cask — is in
[distribution.md](distribution.md); copy [`goreleaser.sample.yml`](goreleaser.sample.yml) and keep
`builds.binary` **plain** so the cask installs as `<tool>`.

## 6. The update hint

In your root command's `PersistentPostRunE`, run
`updatecheck.Hinter{Repo, Bin, Module, Current, CacheFile}.Print(ctx, w)` (gated on `dev` builds +
an opt-out env). Users get the daily "newer version — run: `<upgrade command>`" line for free.

## Conventions

Inherited via the synced `.claude/rules`: conventional commits, TDD, the `charm.land` registry,
SHA-pinned actions, exact version pins. See [`docs/rules/`](../rules/).
