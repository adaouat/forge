---
name: new-tool
description: Use when bootstrapping/scaffolding a brand-new adaouat/* CLI (the bifrost/heraut family) on forge — creates the tool's Tier-2 scaffolding, the cli.Run main, its accent, CI, release wiring, and the update hint. Semi-automated: runs the deterministic steps but pauses at the judgment points (accent hue, coverage %, command tree).
---

# Bootstrap a new adaouat/* CLI on forge

This is the executable form of `docs/guides/new-tool.md` (read it for the *why*). It creates a new
`github.com/adaouat/<tool>` CLI as a **sibling of forge** — at `../<tool>` relative to this repo.
forge does the heavy lifting; the tool wires a thin `main`.

**Semi-automated.** Run the deterministic steps yourself. **Stop and ask** at the three judgment
points — never guess them:

| # | Judgment point | Why it's yours |
|---|---|---|
| Step 3 | **Accent hue** | Brand. Taken: bifrost teal/violet, heraut gold/azure, forge Ember (orange/coal-red). Pick a *distinct* hue, or inherit Ember by passing a zero `Accent`. |
| Step 4 | **Coverage threshold** | `go-ci.yml` requires it, no default — it's per-tool policy. |
| Step 2 | **The command tree** | What the root command does. Leave a stub if it isn't decided yet. |

## 0. Inputs & guards

- `<tool>` — lowercase kebab; the binary name **and** module path (`github.com/adaouat/<tool>`).
- **Abort** if `../<tool>` already exists — don't overwrite.
- Resolve the **latest forge tag + its SHA** now (you'll pin CI to it):
  `git tag --sort=-v:refname | head -1` and `git rev-list -n1 <tag>` in this repo. At time of
  writing that's `v0.8.0` / `5583d84` — re-resolve, don't hardcode.

## 1. Scaffolding first (Tier-2) — this installs the pinned Go

Do this **before** any Go. `.config/mise` is what installs the pinned Go (and golangci-lint, hk),
so it must come first. Run from the workspace root (forge's parent):

```bash
mkdir ../<tool> && git -C ../<tool> init
cp -R .config ../<tool>/.config         # mise, hk, cocogitto, typos, yamlfmt — pure tooling, identical
mkdir -p ../<tool>/.claude/rules
cp docs/rules/workflow.md docs/rules/testing.md docs/rules/coding.md ../<tool>/.claude/rules/
cp docs/rules/agent.md ../<tool>/.claude/rules/claude.md
cat > ../<tool>/.gitignore <<'EOF'
# Build outputs
/<tool>
*.test
/dist
/coverage.out

# IDE
/.idea

# Mise
/.config/mise/config.local.toml

# Claude Code — local-only permissions
/.claude/settings.local.json
EOF
```

> **The `.gitignore` is required, not cosmetic** — hk's `yamlfmt` step aborts the whole lint run
> with `gitignore not found` if it's missing. (Remember to expand `<tool>` inside the heredoc.)

Then:
- **Adapt `.config/heraut.yml`** — change `repository: adaouat/forge` → `adaouat/<tool>` (on macOS,
  `sed -i ''`; GNU `sed -i`).
- **PAUSE — adapt the rules.** forge's rules are *library*-flavored ("zero domain logic", the
  extraction bar). A tool has domain logic — replace that framing. Keep the shared conventions
  (conventional commits, TDD, `charm.land` registry, SHA-pinned actions, version pins). Use
  `../bifrost/.claude/rules` or `../heraut/.claude/rules` as tool-flavored examples.
- Write `../<tool>/CLAUDE.md` importing them (and the tool's own preamble):
  ```
  @.claude/rules/workflow.md
  @.claude/rules/testing.md
  @.claude/rules/coding.md
  @.claude/rules/claude.md
  ```
- Install the toolchain (needs the git repo from above — `hk install` runs on postinstall):
  ```bash
  cd ../<tool> && mise trust && mise install
  ```

Now `go`, `golangci-lint`, hk are the exact family versions. Run later `go`/`hk` from inside the
tool dir (mise auto-detects `.config/mise/config.toml`), or prefix `mise exec --`.

## 2. Bootstrap the module

Run from inside `../<tool>` — and run each `go`/`mise`/`hk` command as `cd ../<tool> && <cmd>` in a
**single** invocation: the working directory does **not** persist between separate shell calls here.

```bash
cd ../<tool> && go mod init github.com/adaouat/<tool>
cd ../<tool> && go get github.com/adaouat/forge@latest
```

> `go get` may bump `go.mod`'s Go line past the mise pin (forge needs ≥ the latest patch; the mise
> config pins the minor, e.g. `go = "1.26"`). Go's toolchain manager fetches the exact patch
> transparently, so builds work — the `go.mod` patch floating ahead of the mise pin is expected.

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

`cmd/<tool>/root.go` (**PAUSE for the command tree** — wire real subcommands, or leave the stub):

```go
package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/adaouat/forge/updatecheck"
)

func rootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "<tool>",
		Short:         "<one-line description>",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentPostRunE = func(c *cobra.Command, _ []string) error {
		updateHint(c, version) // step 6
		return nil
	}
	// cmd.AddCommand(...)  ← your subcommands
	return cmd
}
```

**fang is reached only through `cli.Run`** — don't import it directly (its version stays forge's).
cobra you *do* import to build commands; keep its version on forge's pin.

## 3. Pick an accent — PAUSE for the hue

`internal/ui/theme.go`:

```go
package ui

import (
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	forgeui "github.com/adaouat/forge/ui"
)

// Accent is <tool>'s brand over forge's shared palette — <NAME>, <hue rationale>.
func Accent() forgeui.Accent {
	return forgeui.Accent{
		Light:          lipgloss.Color("#<primary-light>"),   // titles / program / flags
		Dark:           lipgloss.Color("#<primary-dark>"),
		SecondaryLight: lipgloss.Color("#<secondary-light>"), // subcommands
		SecondaryDark:  lipgloss.Color("#<secondary-dark>"),
	}
}

// HuhTheme themes interactive prompts to match. Use via form.WithTheme(ui.HuhTheme()).
func HuhTheme() huh.ThemeFunc { return forgeui.HuhTheme(Accent()) }
```

To inherit Ember instead, return `forgeui.Accent{}`. The shared structure (args blue, muted gray,
errors red) comes from the palette either way — the tool still reads as family.

## 4. CI — PAUSE for the coverage threshold

`.github/workflows/ci.yml` (pin `go-ci.yml` to the forge SHA from step 0):

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:
jobs:
  ci:
    uses: adaouat/forge/.github/workflows/go-ci.yml@<forge-sha> # <forge-tag>
    with:
      coverage-threshold: <PCT>   # required — no default
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6
      - name: Setup Mise
        uses: jdx/mise-action@1648a7812b9aeae629881980618f079932869151 # v4
      - name: Build
        run: go build ./...
```

## 5. Release + distribution

heraut releases the whole family. This part is template-copy, not generated — point at the guides:

- **`.goreleaser.yml`** — copy `docs/guides/goreleaser.sample.yml`, replace every `<app>` with
  `<tool>`, fill the description. Keep `builds.binary` **plain** so the Homebrew cask installs as
  `<tool>`. Decide release ownership (`release.disable` true → heraut-owned, needs `url.template`).
- **`.config/heraut.yml`** — already adapted in step 1.
- **`.github/workflows/release.yml`** — copy `../heraut/.github/workflows/release.yml` (or
  bifrost's): it calls forge's **release-setup** composite (ADR-0009), then `heraut release`, plus
  the goreleaser build + artifacts.json collect + cask-push steps. forge's *own* release.yml omits
  those (forge is a library) — don't use it as the tool template.

Full model + the build-only/Homebrew variants: `docs/guides/distribution.md`.

## 6. The update hint

`cmd/<tool>/root.go`, referenced by `PersistentPostRunE` above:

```go
func updateHint(c *cobra.Command, version string) {
	if version == "dev" || os.Getenv("<TOOL>_CHECK_UPDATE") == "false" {
		return
	}
	cache, _ := os.UserCacheDir()
	updatecheck.Hinter{
		Repo:      "adaouat/<tool>",
		Bin:       "<tool>",
		Module:    "github.com/adaouat/<tool>/cmd/<tool>",
		Current:   version,
		CacheFile: filepath.Join(cache, "<tool>", "update-check.json"),
	}.Print(c.Context(), c.ErrOrStderr())
}
```

Users get the daily "newer version — run: `<upgrade command>`" line for free (silent on `dev`).

## 7. Verify, then make the first commit

Run each from inside `../<tool>` (`cd ../<tool> && …` per call — cwd doesn't persist between calls):

```bash
cd ../<tool> && go mod tidy
cd ../<tool> && go build ./...
cd ../<tool> && go test ./...
cd ../<tool> && mise run lint:check
```

All green → **make the first commit** so the tool is immediately workable, not a pile of untracked
files. The hk hooks installed in step 1 apply now — `pre-commit` lint and `commit-msg`
conventional-commit + typos — so this also proves they fire:

```bash
cd ../<tool> && git add -A && git commit -m "chore: bootstrap <tool> on forge"
```

(A freshly-scaffolded tool is lint-clean and the message is conventional, so this passes; a
non-conventional message is rejected by the hook.) **Stop there** — don't push, don't tag, don't cut
a release; those are the user's call. Then summarize the PAUSE decisions (accent, coverage %, command
tree) and what's left to flesh out.

## References

`docs/guides/new-tool.md` (the prose), `tier2-sync.md`, `distribution.md`,
[ADR-0007](docs/adr/0007-public-api-surface-and-stability.md) (the contract),
[ADR-0009](docs/adr/0009-release-setup-composite-action.md),
[ADR-0010](docs/adr/0010-cli-framework-foundation.md) (forge as the framework foundation).
