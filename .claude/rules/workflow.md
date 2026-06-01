# Workflow rules

## Branching

**During the build phase (pre-v1.0)**: commits land directly on `main`. The roadmap is the
protection, not branches. One developer, one trunk.

**After v1.0 ships**: every working session starts on a new branch off `main`. Never commit
directly to `main`.

- Branch name (post-v1.0): `<type>/<short-description>` where type matches the
  conventional-commit type (e.g. `feat/exec-runner`, `fix/path-precedence`).
- Fetch with prune before branching:
  ```bash
  git fetch --prune --prune-tags --all --tags
  git checkout -b <type>/<short-description> origin/main
  ```

## Conventional commits

All commits follow [Conventional Commits](https://www.conventionalcommits.org/). Allowed types:

| Type       | Use for                                                    |
|------------|------------------------------------------------------------|
| `feat`     | New exported behaviour in a forge package                  |
| `fix`      | Bug fix in existing behaviour                              |
| `docs`     | `docs/specs/`, `docs/adr/`, README, in-code doc comments   |
| `chore`    | Tool config, repo housekeeping, dependency bumps           |
| `refactor` | Code change with no behaviour change                       |
| `test`     | Adding or rewriting tests, no production change            |
| `style`    | Formatting, whitespace, lint-only fixes                    |
| `perf`     | Performance-only change                                    |
| `ci`       | `.github/workflows/*`, release tooling                     |
| `build`    | `go.mod`, build system                                     |

**Scope** matches the affected package: `feat(exec): add RunEnv`, `fix(config): env var
beats .config path`, `docs(adr): add 0002 exec runner contract`. Keep subject lines ≤72
characters. Use the body for the *why*, not the *what*.

## Two-step roadmap flow

Task status is tracked inline in `docs/tasks/roadmap.md` via `[ ]` / `[x]` checkboxes.

1. **Implement** — confirm the task is `[ ]`, then do the work (TDD: failing test first).
   Commit in logical pieces using the right conventional-commit type.
2. **Complete** — flip `[ ]` → `[x]` and add a one-paragraph note under the task describing
   actual decisions, deferred items, or deviations. Commit the roadmap update alongside the
   final implementation commit.

Never silently mark a task complete without the note. The note is what makes the roadmap a
living document.

## Git hooks (hk)

Hooks live in `.config/hk/config.pkl` and run on every commit (pre-commit linters,
commit-msg conventional-commit validation, prepare-commit-msg `typos`).

**Never** pass `--no-verify`, `--no-gpg-sign`, or any flag that bypasses hooks. If a hook
fails, fix the underlying issue.

## Lint fixes

Fix lint failures through `hk`, never the underlying tool directly (it applies the project's
configured file selection and flags):

```bash
hk fix             # fix everything fixable
hk fix -S <linter> # target one linter (e.g. hk fix -S golangci-lint, hk fix -S yamlfmt)
```

## Version pinning

Pin exact versions everywhere — no `latest` in mise config, `go.mod`, or CI workflows.

**Exceptions** — format/lint or editor tooling with no API surface that could break the
build may use `latest`: `pkl`, `tombi`, `typos`, `yamlfmt`, `gopls` (and similar LSP tools).

## GitHub Actions

Pin every action to a full commit SHA, never a mutable tag. Add the semantic version as a
comment so intent stays readable:

```yaml
uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6
```

To update, find the new SHA for the desired tag (`github.com/<owner>/<action>/tags`) and
replace both the SHA and the comment. Never use `@v4`, `@main`, or `@latest`.

## Plans

Plans live in `.claude/plans/`. Each captures one discrete unit of work — a phase, a
milestone, a non-trivial task, or a research/design spike. Name them descriptively in
lowercase kebab-case (`m1-exec-extraction.md`); never keep the auto-generated random name.

## Releases

forge is a **library** — there is no binary and no GoReleaser. A release is a `v*` git tag;
consumers pick it up by bumping their `go.mod` and running `go mod tidy`. A tag that changes
an exported contract is preceded by the ADR that justifies it.
</content>
