# ADR-0003 — Shared exit-code vocabulary

**Status:** Accepted
**Date:** 2026-06-02

## Context

M2.1 extracted the exit-code *mechanism* (`ExitError`, `Wrap`, `Resolve`) into forge but
deliberately shipped **no exit-code value constants**, reasoning that heraut's
`Config`/`Runtime`/`Promotion` were domain and bifrost "used raw ints".

That was too conservative. Mapping the two apps shows the *values and their meanings* are
already shared, bifrost simply never named them:

| Code | meaning | heraut | bifrost |
|---|---|---|---|
| `0` | success | `OK` | implicit |
| `1` | bad flags/args; unclassified default | `Usage` | raw `1` |
| `2` | invalid config / validation failure | `Config` | raw `2` (≈6 sites) |
| `3` | external command / network / IO failure | `Runtime` | raw `3` (≈5 sites) |
| `4` | promotion guard tripped | `Promotion` | — (domain, heraut-only) |
| `70` | unexpected internal condition | `Internal` | — (heraut-only) |

`0/1/2/3` are used **identically** in both apps; bifrost's bare `2`/`3` literals are exactly
the magic numbers named constants exist to remove. A generic exit-code vocabulary is CLI
process plumbing, not domain logic, and a third family CLI is imminent — standardising now
means it starts consistent rather than inventing its own scheme.

## Decision

forge's `exitcode` package defines the **generic** exit-code vocabulary:

```go
const (
    OK       = 0  // success
    Usage    = 1  // bad flags/args; default for unclassified errors
    Config   = 2  // invalid config / validation failure
    Runtime  = 3  // external command, network, or IO failure
    Internal = 70 // unexpected internal condition (sysexits EX_SOFTWARE)
)
```

`Resolve` returns `OK` for nil and `Usage` for an unclassified error. `Internal` is included
as a generic convention even though only heraut uses it today; bifrost/the next tool adopt it
for unexpected failures.

**Domain codes stay in the apps.** heraut keeps `Promotion = 4` in its own package; forge
never learns release-management semantics.

**Extension convention:** forge owns `0–3` and `70`. Apps define their own domain codes in
the reserved range **`4–69`** (heraut: `Promotion = 4`), layered on top of forge's set via
their local facade (`heraut/internal/exitcode`, `bifrost/internal/cmderr`).

This supersedes the "no value constants in forge" note in roadmap M2.1.

## Consequences

- **Family-wide consistent exit codes.** A script wrapping any `adaouat/*` CLI can rely on
  `2 = config error`, `3 = runtime error` regardless of which tool it ran — a real UX win and
  the point of a shared foundation.
- **The numbering is now a contract.** Changing a value is a coordinated breaking change
  across every consumer; that is the intended weight, and why it is recorded here.
- **bifrost's magic numbers become named**, and the next tool starts on the shared vocabulary.
- Each app re-exports the generic codes through its existing facade and adds only its own
  domain codes, so call sites stay terse (`exitcode.Config`, `cmderr.Config`).
