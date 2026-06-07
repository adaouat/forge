---
description: Bump the consumer apps (heraut, bifrost) to a forge version — go.mod only, build/test/lint, commit
argument-hint: [version]
---

Bump every consumer of forge to **$ARGUMENTS** (a tag like `v0.9.0`; if omitted, the latest forge tag).

Invoke the **`bump-forge` skill** and follow it exactly. It is **semi-automated**: it bumps `go.mod`,
runs build/test/lint, and commits **`go.mod`/`go.sum` only** in each app (never `git add -A`, so WIP
like `.goreleaser.yml` is preserved) — but it **stops and asks** at the two judgment points (the commit
type/message, and whether to push). If no version was given, resolve the latest forge tag first.
