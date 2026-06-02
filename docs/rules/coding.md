# Coding rules

## What belongs in forge

forge is a **library**, imported by ≥2 repos. Before writing anything, check it against the
extraction bar in [ADR-0001](../../docs/adr/0001-shared-core-module.md): identical across
consumers + stable contract + ≥2 real consumers. **Zero domain logic** lives here.

Current packages and their sources are tracked in `docs/tasks/roadmap.md`. Do not add a new
package without a roadmap task and (for a load-bearing interface) an ADR.

## Public API is a contract

Because two repos import forge, every exported identifier is a contract, not an internal
detail.

- Keep the exported surface **minimal and deliberate**. Unexport anything a consumer does
  not need.
- A breaking change to an exported signature needs an ADR and a coordinated bump of both
  consumers — never a silent edit.
- **Accept interfaces, return structs.** Signatures take the narrowest interface that works
  and never mention an app-specific type.
- No app-specific names, constants, or assumptions leak into forge. Parameterize instead
  (e.g. config path resolution takes the app name; it does not hardcode `heraut`).

## Error handling

- **Always wrap.** Every `if err != nil` returns `fmt.Errorf("doing X: %w", err)`. The `%w`
  is mandatory; without it, callers cannot `errors.Is` / `errors.As`.
- **Never string-match errors.** Use `errors.Is(err, target)` for sentinels and
  `errors.As(err, &typed)` for typed errors.
- **Expose sentinels/typed errors at package boundaries** so consumers can classify failures
  (e.g. `exitcode` maps them to process codes).
- **Never panic in library code** — return an error.
- **Never call `os.Exit` anywhere in forge.** Forge is imported; only the consuming app's
  `main` decides the process exit code (that is what the `exitcode` package is for).

## Code quality

- No comments unless the *why* is non-obvious — a hidden constraint, a subtle invariant, a
  workaround for a specific bug. Never describe *what* the code does; well-named identifiers
  do that. Reference the ADR/bug when relevant.
- No multi-paragraph docstrings. (Exported identifiers still get a one-line doc comment —
  it is a public library.)
- No features, abstractions, or refactoring beyond what the current task requires. YAGNI.
- Three similar lines are better than a premature abstraction.
- Only validate at boundaries: caller input, external APIs, config files. Trust internal
  guarantees — no error handling for scenarios that cannot happen.
- No backwards-compatibility shims for code that does not exist yet.

## Charmbracelet dependencies

All charmbracelet packages use the `charm.land` module registry, not
`github.com/charmbracelet`.

```
go get charm.land/<module>/v2   # e.g. charm.land/huh/v2, charm.land/bubbles/v2
```

Never add `github.com/charmbracelet/<module>` as a direct dependency.
</content>
