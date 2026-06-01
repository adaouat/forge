# Testing rules

## TDD is required

Write the failing test before writing implementation code. The cycle is **red → green →
refactor**. If you are tempted to skip the test "because the change is small", stop — that
is exactly when the test is most valuable.

If the user asks you to skip tests, push back and explain why. Do not silently agree.

## Coverage discipline

Every exported function in core has tests — core is consumed by two repos, so an untested
edge case becomes two repos' bug. Each package ships with the tests that already covered it
in its source app, ported, plus any gaps closed.

## Table-driven tests preferred

Group related cases into one test with a `[]struct` of inputs and expected outputs. Each row
gets a descriptive `name`:

```go
tests := []struct {
    name string
    in   string
    want string
}{
    {"flag wins over env", ...},
    {"env wins over .config", ...},
}
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) { /* … */ })
}
```

## `exectest.MockRunner` — the contract-test workhorse

core ships `exec/exectest` so consumers (and core's own tests) can assert *which CLI args
were passed* without touching real exec. `MockRunner` implements `exec.Runner`, queues
responses FIFO, and records every `Call`:

```go
mr := exectest.NewMockRunner()
mr.QueueResponse("", "", nil) // stdout, stderr, err

// … exercise code that takes an exec.Runner …

require.Len(t, mr.Calls, 1)
assert.Equal(t, "git", mr.Calls[0].Name)
assert.Equal(t, []string{"tag", "v1.2.3"}, mr.Calls[0].Args)
```

When the assertion is about which args were passed, use `MockRunner`. Never reach for real
exec. Use `exectest.FakeBin` only when the test genuinely needs the real exec path
(stdin/stdout forwarding, env propagation, exit-code mapping) — sparingly.

## Determinism — never break these

- **No time-of-day dependencies.** Any time-sensitive code takes a `now func() time.Time` so
  tests can fix the clock.
- **No network calls.** Self-update / HTTP code is tested against an `httptest.Server`,
  never the real GitHub/GitLab API.
- **No filesystem outside `t.TempDir()`.** Config and path tests write into a temp dir; they
  never touch the source tree or the real home directory.
- **No environment leakage.** Set env with `t.Setenv(...)` so it is restored on test exit —
  never `os.Setenv` directly.

## Preserve hard-won edge cases

The suite encodes hard-won edge cases (path-resolution precedence, version arithmetic like
`v1.9.0` → `v1.10.0`, SHA-256 verification failure paths, …). **Never delete a test row to
make a change easier.** An assertion is load-bearing until proven otherwise. Drop a row only
when the behaviour it tested is deliberately changed — and only with an ADR documenting it.

## Shared helpers live in one place

Reusable test helpers go in the package that owns the contract (`exec/exectest`, etc.), never
duplicated across test files. If a second test needs the same setup, move it to the shared
helper first, then both call it.

## When a hook or test fails

Fix the root cause. Do not comment out the assertion, add an unexplained `t.Skip()`, loosen
the assertion, or suppress the linter. Each defeats the test's purpose. If the test itself is
wrong, fix it in a separate commit with an explanation.
</content>
