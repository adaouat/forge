# Guides

How-to documentation for the `github.com/adaouat/*` CLI family — conventions an app
follows, copied and adapted rather than imported (forge is a library).

| Guide | Covers |
|---|---|
| [`distribution.md`](distribution.md) | Build / publish / install model — goreleaser raw-binary format, release ownership, mise / curl / Homebrew channels. Template: [`goreleaser.sample.yml`](goreleaser.sample.yml). |
| [`tier2-sync.md`](tier2-sync.md) | How an app refreshes its `.claude/rules` + `.config` from forge's canonical `docs/rules` + `.config`. |
| [`new-tool.md`](new-tool.md) | Start a new CLI on forge — `cli.Run` + accent, scaffolding, CI, release, the update hint. |
