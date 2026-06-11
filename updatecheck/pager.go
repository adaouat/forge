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
// the pager can't be started, it returns false so render falls back to writing
// content directly to w — paging must never cause whatsnew to fail or swallow
// output. Once the pager starts, its exit status is irrelevant to the caller: a
// non-zero exit (e.g. the user pressing Ctrl-C in less) must not cause render to
// also print the raw content afterward. See ADR-0012's 2026-06-11 refinement.
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

	if err := cmd.Start(); err != nil {
		return false
	}
	_ = cmd.Wait()
	return true
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
