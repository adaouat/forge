# Architecture Decision Records

ADRs document significant architectural choices: what was decided, why, and what trade-offs were accepted.

| ADR | Title | Status |
|---|---|---|
| [0001](0001-shared-core-module.md) | Extract the shared `forge` module | Accepted |
| [0002](0002-exec-runner-working-directory.md) | `exec.Runner` gains a working-directory method | Accepted |
| [0003](0003-shared-exit-code-vocabulary.md) | Shared exit-code vocabulary | Accepted |
| [0004](0004-ui-spinner-task-runner.md) | `ui.Spinner` task runner | Accepted |
| [0005](0005-updates-via-package-managers.md) | Updates via package managers, not binary self-replacement | Accepted |
| [0006](0006-shared-ci-reusable-workflow.md) | Shared lint/test CI via a reusable workflow in forge | Accepted |
| [0007](0007-public-api-surface-and-stability.md) | Public API surface and stability contract | Accepted |
| [0008](0008-ui-theme-palette.md) | Family UI theme: shared palette, per-tool accent | Accepted |
| [0009](0009-release-setup-composite-action.md) | Shared release setup via a composite action | Accepted |
| [0010](0010-cli-framework-foundation.md) | forge is the CLI framework foundation | Accepted |
| [0011](0011-logging-foundation.md) | Logging foundation: `slog` API, `charmbracelet/log` handler | Accepted |
| [0012](0012-whatsnew-changelog.md) | `whatsnew`: in-terminal changelog after the update check | Accepted |
