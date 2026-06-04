# Syncing Tier-2 scaffolding (rules + tooling)

forge is the canonical source for the **Tier-2 scaffolding** the family shares — the
conventions in [`docs/rules/`](../rules/) and the `.config/` tooling baseline. Unlike the
Tier-1 Go packages (imported via the module), this scaffolding can't be a live dependency:
prose rules and tool-config files aren't `import`able, and the OSS tools have no remote-include.
So each app keeps an **adapted copy**, and refreshing it from forge is a deliberate,
diff-and-apply step — this guide.

## What forge canonicalizes

| forge (canonical) | app copy | Notes |
|---|---|---|
| `docs/rules/workflow.md` | `.claude/rules/workflow.md` | conventional commits, branching, hk, releases |
| `docs/rules/testing.md` | `.claude/rules/testing.md` | TDD, table-driven, determinism |
| `docs/rules/coding.md` | `.claude/rules/coding.md` | error handling, `charm.land`, version pinning |
| `docs/rules/agent.md` | `.claude/rules/**claude.md**` | renamed on the app side (historical) |
| `.config/mise/` | `.config/mise/` | shared tool **pins** (go, golangci-lint, hk, …) |
| `.config/hk/` | `.config/hk/` | linter set (golangci-lint, yamlfmt, typos, actionlint) |
| `.config/{cocogitto,typos,yamlfmt}/` | same | commit-lint + format config |

The apps' copies are **adapted, not identical**: heraut's `coding.md` is ~2× forge's (hexagonal
layers, pipeline rules); bifrost's are leaner; both add app-specific tools to `.config/mise`
(goreleaser, and Docker/hadolint for heraut). The shared *baseline* tracks forge; the
app-specific parts stay in the app.

## Not synced (app-owned)

App-specific rules (bifrost's deploy/atomic, heraut's pipeline/generators/versioning),
app-specific tools (`goreleaser`, Docker), and per-app config (`.goreleaser.yml`,
`.config/heraut.yml`, app `cliff.toml`). forge holds none of these — see
[ADR-0001](../adr/0001-shared-core-module.md) Tier 3.

## The sync, step by step

forge and the apps are local siblings (`../forge`), so diff against the canonical copy:

```bash
# rules — mind the agent.md ↔ claude.md rename
diff -u ../forge/docs/rules/workflow.md .claude/rules/workflow.md
diff -u ../forge/docs/rules/coding.md   .claude/rules/coding.md
diff -u ../forge/docs/rules/testing.md  .claude/rules/testing.md
diff -u ../forge/docs/rules/agent.md    .claude/rules/claude.md

# tooling
diff -ru ../forge/.config/hk   .config/hk
diff -ru ../forge/.config/mise .config/mise   # ignore app-only tool lines
```

Apply forge's changes **by hand**, keeping the app-specific sections — these files are adapted
prose/config and don't mechanically merge. The most mechanical case is a **shared tool-version
bump**: align the pin in `.config/mise/config.toml` to forge's, then `mise install` + regenerate
`mise.lock`.

## Direction and cadence

- **forge is upstream.** Changes flow forge → apps. If an app discovers a better convention,
  **promote it to forge first** (edit `docs/rules/`), then sync it down — so forge stays
  canonical and the next tool inherits it.
- **Deliberate, not automatic.** Like the SHA-pinned reusable-workflow refs, an app refreshes
  when forge changes something shared — there's no obligation to track every forge commit. A
  good trigger is a forge release that touches `docs/rules/` or `.config/`.
