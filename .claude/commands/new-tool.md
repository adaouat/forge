---
description: Bootstrap a new adaouat/* CLI on forge (semi-automated scaffold)
argument-hint: <tool-name>
---

Bootstrap a new `github.com/adaouat/*` CLI named **$ARGUMENTS** on forge.

Invoke the **`new-tool` skill** and follow it exactly — it is the executable form of
`docs/guides/new-tool.md`. The new tool is created as a sibling of forge (`../$ARGUMENTS`).

It is **semi-automated**: run the deterministic steps yourself, but stop and ask at the three
judgment points (accent hue, coverage threshold, the command tree). It ends by making the first
commit (`chore: bootstrap <name> on forge`) so the tool is workable immediately — but never pushes,
tags, or releases. If no name was given above, ask for one before starting.
