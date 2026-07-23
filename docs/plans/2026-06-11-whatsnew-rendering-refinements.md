# `whatsnew` Rendering Refinements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix `whatsnew`'s silent fallback to raw, unstyled markdown (caused by glamour's
terminal-querying `"auto"` style) and add `$PAGER`/`$NO_PAGER`-based paging, per
[the approved design](../specs/2026-06-11-whatsnew-rendering-design.md).

**Architecture:** Both changes live entirely in `render()` (`updatecheck/whatsnew.go`) and a
new `updatecheck/pager.go`. Style selection becomes a pure function of `ui.HasColor(w)` (no
terminal queries). After rendering, `render()` tries `pagedOutput(w, out)` (TTY + resolvable
`$PAGER`); if that returns `false` it writes `out` to `w` directly, exactly as today.

**Tech Stack:** Go, `charm.land/glamour/v2`, `github.com/adaouat/forge/ui`
(`HasColor`/`IsTTY`), `os/exec`, `testify`.

---

### Task 1: Roadmap + ADR doc updates

**Files:**
- Modify: `docs/tasks/roadmap.md` (add M12 section after M11, before "Explicitly NOT on this
  roadmap")
- Modify: `docs/adr/0012-whatsnew-changelog.md` (add a new "Refinement" section after the
  existing "Refinement (decided at C, 2026-06-10)" section, before "### Phasing → roadmap")

- [ ] **Step 1: Add the M12 section to the roadmap**

Open `docs/tasks/roadmap.md` and find the line `## Explicitly NOT on this roadmap` (currently
line 806). Insert the following section **immediately before** that line (keep the existing
blank line separating M11 from it):

```markdown
## M12 — `whatsnew` rendering refinements *([ADR-0012](../adr/0012-whatsnew-changelog.md) refinement)*

*`whatsnew`'s `render()` used glamour's `"auto"` style, which queries the terminal for its
background color via OSC11 and silently falls back to raw, unstyled markdown when the
terminal doesn't answer — observed in real use (heraut `whatsnew` under a pty showed 4
unanswered OSC11 queries then raw markdown). ADR-0012 also deferred a pager. Both are
revisited here; see the ADR's 2026-06-11 refinement note.*

- [ ] **Style selection via `ui.HasColor`, drop `"auto"`** — `render()` picks glamour style
      `"dark"` when `ui.HasColor(w)`, else `"notty"`, with no terminal query. TDD:
      `TestGlamourStyle` (buffer with no color → `"notty"`; `CLICOLOR_FORCE=1` + `TERM` set →
      `"dark"`). Existing `TestRender` stays green (non-TTY buffer → `"notty"`, same
      observable output as before).
- [ ] **Pager via `$PAGER`/`$NO_PAGER`** — when `ui.IsTTY(w)`, pipe rendered output through
      `$PAGER` (default `less`, `LESS=FRX` if `$LESS` unset, git's convention: `-F` exit if
      content fits one screen, `-R` show ANSI colors, `-X` don't clear screen on exit);
      skipped via `$NO_PAGER` or when no pager binary is found, falling back to writing
      directly to `w` on any resolution/spawn failure. New `updatecheck/pager.go`
      (`resolvePager`, `pagedOutput`), TDD via `TestResolvePager` (table-driven, fake `$PATH`
      stubs) and `TestPagedOutput_NonTTY`. No new exported surface (ADR-0007 unaffected).
```

- [ ] **Step 2: Add the ADR-0012 refinement section**

Open `docs/adr/0012-whatsnew-changelog.md`. Find the line `### Phasing → roadmap` (currently
line 133). Insert the following section **immediately before** that line (keep the existing
blank line separating it from the prior refinement section):

```markdown
### Refinement (decided 2026-06-11): drop terminal-query styling, add a pager

Real-world use surfaced a failure mode in C's `render()`: `glamour.Render(md, "auto")`
queries the terminal for its background color via OSC11 escape sequences to pick a
light/dark style. When the terminal doesn't answer (reproduced under a pty via `script`: 4
unanswered OSC11 queries), glamour times out and `render()`'s existing fallback
(`out = md`) emits **raw, unstyled markdown** — the reported "black/white, missing styling"
behavior. Resolved by dropping `"auto"` entirely: `render()` now picks glamour style
`"dark"` when `ui.HasColor(w)` (forge's existing `NO_COLOR`/`CLICOLOR_FORCE`/`TERM=dumb`/TTY
check, no terminal round-trip) and `"notty"` otherwise — `"notty"` still formats markdown
structurally without raw `#`/`**` syntax, strictly better than the old fallback. No new
light/dark heuristic: a dark-style render on a light background is still readable, and most
dev terminals default to dark themes.

Separately, this also lifts the pager deferral ("Add one only if changelog length warrants
it" — Tier D's multi-release spans plus the embedded-changelog fallback can be long).
`render()` now pipes its output through `$PAGER` (default `less`, with `LESS=FRX` set when
`$LESS` is unset) when `w` is a TTY, opt-out via `$NO_PAGER`, matching `git`/`gh`
conventions — no new flags, no new forge dependency. Falls back to direct output on any
resolution or spawn failure, so paging can never cause `whatsnew` to fail or swallow output.

Full design: `docs/specs/2026-06-11-whatsnew-rendering-design.md`.
```

- [ ] **Step 3: Commit**

```bash
git add docs/tasks/roadmap.md docs/adr/0012-whatsnew-changelog.md
git commit -m "docs(roadmap): add M12 whatsnew rendering refinements"
```

---

### Task 2: Style selection via `ui.HasColor`, drop `"auto"`

**Files:**
- Modify: `updatecheck/whatsnew.go`
- Modify: `updatecheck/whatsnew_test.go`
- Modify: `docs/tasks/roadmap.md` (flip M12's first task to `[x]` with a note)

- [ ] **Step 1: Write the failing test**

In `updatecheck/whatsnew_test.go`, add this test (place it right after `TestRender`):

```go
func TestGlamourStyle(t *testing.T) {
	t.Run("no color support -> notty", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		t.Setenv("CLICOLOR_FORCE", "")
		assert.Equal(t, "notty", glamourStyle(&bytes.Buffer{}))
	})

	t.Run("color forced -> dark", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		t.Setenv("CLICOLOR_FORCE", "1")
		t.Setenv("TERM", "xterm-256color")
		assert.Equal(t, "dark", glamourStyle(&bytes.Buffer{}))
	})
}
```

`bytes` is already imported in this file.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./updatecheck/... -run TestGlamourStyle -v`
Expected: FAIL — `undefined: glamourStyle`

- [ ] **Step 3: Implement `glamourStyle` and wire it into `render`**

In `updatecheck/whatsnew.go`, add the import and the new function, and update `render`:

```go
import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"charm.land/glamour/v2"
	"github.com/adaouat/forge/ui"
	"github.com/spf13/cobra"
)
```

```go
// glamourStyle picks a glamour style for w without querying the terminal: "dark" when w
// supports color (per ui.HasColor), "notty" otherwise. Avoids glamour's "auto" style, which
// queries the terminal via OSC11 and silently falls back to raw markdown when the terminal
// doesn't answer. See ADR-0012's 2026-06-11 refinement.
func glamourStyle(w io.Writer) string {
	if ui.HasColor(w) {
		return "dark"
	}
	return "notty"
}

// render writes md to w through glamour, falling back to the raw markdown if glamour fails —
// the styled render is best-effort, but the content must always reach the user. See ADR-0012.
func render(w io.Writer, md string) error {
	out, err := glamour.Render(md, glamourStyle(w))
	if err != nil {
		out = md
	}
	if _, err := io.WriteString(w, out); err != nil {
		return fmt.Errorf("writing changelog: %w", err)
	}
	return nil
}
```

This replaces the existing `render` function (which called `glamour.Render(md, "auto")`) —
keep the same doc comment, just the body and the new `glamourStyle` helper above it.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./updatecheck/... -v`
Expected: PASS — including `TestGlamourStyle` and the existing `TestRender`,
`TestWhatsNewCommand_RendersSpanNewerThanCurrent`, etc.

- [ ] **Step 5: Flip the roadmap task and add the completion note**

In `docs/tasks/roadmap.md`, change:

```markdown
- [ ] **Style selection via `ui.HasColor`, drop `"auto"`** — `render()` picks glamour style
```

to:

```markdown
- [x] **Style selection via `ui.HasColor`, drop `"auto"`** — `render()` picks glamour style
```

and append this note as a new paragraph directly after that bullet's existing text (before
the next `- [ ]` bullet):

```markdown
      **Done:** `glamourStyle(w io.Writer) string` added to `updatecheck/whatsnew.go`,
      returning `"dark"`/`"notty"` based on `ui.HasColor(w)`; `render` now calls
      `glamour.Render(md, glamourStyle(w))`. No terminal queries. `TestGlamourStyle` covers
      both branches; existing `TestRender` (non-TTY buffer) unchanged in behavior.
```

- [ ] **Step 6: Commit**

```bash
git add updatecheck/whatsnew.go updatecheck/whatsnew_test.go docs/tasks/roadmap.md
git commit -m "feat(updatecheck): drop glamour auto-style for whatsnew

glamour's \"auto\" style queries the terminal for its background color
via OSC11 and silently falls back to raw markdown when the terminal
doesn't answer. Pick \"dark\"/\"notty\" from ui.HasColor instead, with
no terminal round-trip."
```

---

### Task 3: Pager via `$PAGER`/`$NO_PAGER`

**Files:**
- Create: `updatecheck/pager.go`
- Create: `updatecheck/pager_test.go`
- Modify: `updatecheck/whatsnew.go` (wire `pagedOutput` into `render`)
- Modify: `docs/tasks/roadmap.md` (flip M12's second task to `[x]` with a note)

- [ ] **Step 1: Write the failing tests**

Create `updatecheck/pager_test.go`:

```go
package updatecheck

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeStub(t *testing.T, dir, name string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o755))
}

// unsetEnv removes key from the environment for the test, restoring its prior
// value (or leaving it unset) on cleanup. t.Setenv cannot represent "unset".
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, ok := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, old)
		}
	})
}

func TestResolvePager(t *testing.T) {
	dir := t.TempDir()
	writeStub(t, dir, "less")
	writeStub(t, dir, "cat")
	t.Setenv("PATH", dir)

	t.Run("NO_PAGER disables paging", func(t *testing.T) {
		t.Setenv("NO_PAGER", "1")
		t.Setenv("PAGER", "")
		_, _, ok := resolvePager()
		assert.False(t, ok)
	})

	t.Run("default less with LESS unset sets LESS=FRX", func(t *testing.T) {
		t.Setenv("NO_PAGER", "")
		t.Setenv("PAGER", "")
		unsetEnv(t, "LESS")
		path, extraEnv, ok := resolvePager()
		require.True(t, ok)
		assert.Equal(t, "less", filepath.Base(path))
		assert.Equal(t, []string{"LESS=FRX"}, extraEnv)
	})

	t.Run("default less with LESS already set is left alone", func(t *testing.T) {
		t.Setenv("NO_PAGER", "")
		t.Setenv("PAGER", "")
		t.Setenv("LESS", "-X")
		path, extraEnv, ok := resolvePager()
		require.True(t, ok)
		assert.Equal(t, "less", filepath.Base(path))
		assert.Empty(t, extraEnv)
	})

	t.Run("PAGER=cat is used as-is", func(t *testing.T) {
		t.Setenv("NO_PAGER", "")
		t.Setenv("PAGER", "cat")
		path, extraEnv, ok := resolvePager()
		require.True(t, ok)
		assert.Equal(t, "cat", filepath.Base(path))
		assert.Empty(t, extraEnv)
	})

	t.Run("unknown pager disables paging", func(t *testing.T) {
		t.Setenv("NO_PAGER", "")
		t.Setenv("PAGER", "nonexistent-pager")
		_, _, ok := resolvePager()
		assert.False(t, ok)
	})
}

func TestPagedOutput_NonTTY(t *testing.T) {
	var buf bytes.Buffer
	assert.False(t, pagedOutput(&buf, "content"))
	assert.Empty(t, buf.String(), "pagedOutput must not write to w itself")
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./updatecheck/... -run 'TestResolvePager|TestPagedOutput_NonTTY' -v`
Expected: FAIL — `undefined: resolvePager`, `undefined: pagedOutput`

- [ ] **Step 3: Implement `updatecheck/pager.go`**

```go
package updatecheck

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adaouat/forge/ui"
)

// pagedOutput pipes content through a pager when w is a terminal and a pager is
// available, returning true if it did. When w isn't a TTY, no pager is found, or
// the pager fails to run, it returns false so render falls back to writing
// content directly to w — paging must never cause whatsnew to fail or swallow
// output. See ADR-0012's 2026-06-11 refinement.
func pagedOutput(w io.Writer, content string) bool {
	if !ui.IsTTY(w) {
		return false
	}

	path, extraEnv, ok := resolvePager()
	if !ok {
		return false
	}

	cmd := exec.Command(path)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}

	return cmd.Run() == nil
}

// resolvePager resolves the pager command from $PAGER (default "less"). $NO_PAGER set to
// any non-empty value disables paging entirely. When the resolved pager is "less" and $LESS
// is unset, extraEnv sets LESS=FRX — git's convention: -F exits immediately if content fits
// one screen, -R shows ANSI color codes, -X avoids clearing the screen on exit.
func resolvePager() (path string, extraEnv []string, ok bool) {
	if os.Getenv("NO_PAGER") != "" {
		return "", nil, false
	}

	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}

	path, err := exec.LookPath(pager)
	if err != nil {
		return "", nil, false
	}

	if filepath.Base(pager) == "less" {
		if _, set := os.LookupEnv("LESS"); !set {
			extraEnv = []string{"LESS=FRX"}
		}
	}

	return path, extraEnv, true
}
```

- [ ] **Step 4: Wire `pagedOutput` into `render`**

In `updatecheck/whatsnew.go`, update `render` (from Task 2) to try the pager before writing
directly:

```go
// render writes md to w through glamour, falling back to the raw markdown if glamour fails —
// the styled render is best-effort, but the content must always reach the user. When w is a
// terminal, the rendered output is paged via $PAGER (see pagedOutput). See ADR-0012.
func render(w io.Writer, md string) error {
	out, err := glamour.Render(md, glamourStyle(w))
	if err != nil {
		out = md
	}
	if pagedOutput(w, out) {
		return nil
	}
	if _, err := io.WriteString(w, out); err != nil {
		return fmt.Errorf("writing changelog: %w", err)
	}
	return nil
}
```

- [ ] **Step 5: Run the full `updatecheck` test suite**

Run: `go test ./updatecheck/... -v`
Expected: PASS — all tests, including `TestResolvePager`, `TestPagedOutput_NonTTY`,
`TestGlamourStyle`, `TestRender`, and the `WhatsNewCommand`/`WhatsNewConfig` tests (these all
use `bytes.Buffer`, which is never a TTY, so `pagedOutput` returns `false` and behavior is
unchanged).

- [ ] **Step 6: Run the full module suite and lint**

Run: `go build ./... && go test ./... && hk check`
Expected: build succeeds, all tests pass, `hk check` reports no issues. If `hk check` finds
fixable lint issues, run `hk fix` (or `hk fix -S <linter>` for a specific one) per
`docs/rules/workflow.md`.

- [ ] **Step 7: Flip the roadmap task and add the completion note**

In `docs/tasks/roadmap.md`, change:

```markdown
- [ ] **Pager via `$PAGER`/`$NO_PAGER`** — when `ui.IsTTY(w)`, pipe rendered output through
```

to:

```markdown
- [x] **Pager via `$PAGER`/`$NO_PAGER`** — when `ui.IsTTY(w)`, pipe rendered output through
```

and append this note as a new paragraph directly after that bullet's existing text:

```markdown
      **Done:** `updatecheck/pager.go` adds `resolvePager` (env/`$PATH` resolution,
      `LESS=FRX` default) and `pagedOutput` (TTY check + spawn, passthrough
      stdout/stderr, content via stdin); `render` tries `pagedOutput` before writing
      directly. `TestResolvePager` (table-driven, fake `$PATH` stubs) and
      `TestPagedOutput_NonTTY`. Real interactive pager spawn is not unit-tested
      (consistent with `exectest.FakeBin` being a "sparingly" tool); covered by manual
      verification in Task 4.
```

- [ ] **Step 8: Commit**

```bash
git add updatecheck/pager.go updatecheck/pager_test.go updatecheck/whatsnew.go docs/tasks/roadmap.md
git commit -m "feat(updatecheck): page whatsnew output via \$PAGER

When stdout is a terminal, pipe the rendered changelog through \$PAGER
(default \"less\" with LESS=FRX), opt out via \$NO_PAGER. Falls back to
writing directly on any resolution or spawn failure, matching git/gh
conventions. Lifts ADR-0012's pager deferral."
```

---

### Task 4: Manual verification

**Files:** none (verification only)

- [ ] **Step 1: Build heraut against the local forge changes**

```bash
cd /Users/bchatard/Developer/Adaouat/heraut
go mod edit -replace github.com/adaouat/forge=/Users/bchatard/Developer/Adaouat/forge
go mod tidy
go build -o /tmp/heraut ./cmd/heraut
```

- [ ] **Step 2: Verify styled output under a real pty (no terminal queries)**

```bash
script -q /tmp/out_tty.txt env TERM=xterm-256color NO_PAGER=1 /tmp/heraut whatsnew > /dev/null 2>&1
od -c /tmp/out_tty.txt | head -5
```

Expected: the output starts with ANSI escape sequences (`\033[...m`) for the rendered
heading/text — **no** `\033]11;?\a` (OSC11) queries like before. `NO_PAGER=1` keeps this
non-interactive so it doesn't hang waiting for pager input.

- [ ] **Step 3: Verify the pager interactively**

In a real terminal (not through `script`/this session), run:

```bash
/tmp/heraut whatsnew
```

Expected: output opens in `less` (or `$PAGER` if set), styled with color, scrollable;
quitting `less` (`q`) returns to the shell. Then run:

```bash
NO_PAGER=1 /tmp/heraut whatsnew | cat
```

Expected: styled output prints directly without invoking a pager (piped, non-TTY — same as
before this change for redirected output).

- [ ] **Step 4: Revert the temporary `replace` directive**

```bash
cd /Users/bchatard/Developer/Adaouat/heraut
go mod edit -dropreplace github.com/adaouat/forge
go mod tidy
git status
```

Confirm `go.mod`/`go.sum` are back to their original state (the `replace` was only for local
verification — forge ships this as a tagged release for heraut/bifrost to consume normally).
