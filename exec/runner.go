// Package exec runs external CLI commands behind a small interface so callers
// can swap a recording mock in tests.
package exec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	os_exec "os/exec"
	"strings"
)

// Runner executes external CLI commands.
type Runner interface {
	// Run executes name with args, returning captured stdout and stderr.
	Run(name string, args ...string) (string, string, error)
	// RunEnv executes name with args, appending env to the current process environment.
	RunEnv(env []string, name string, args ...string) (string, string, error)
	// RunDir executes name with args in dir (empty = current dir), appending env.
	RunDir(dir string, env []string, name string, args ...string) (string, string, error)
}

var _ Runner = (*CmdRunner)(nil)

// CmdRunner executes external CLI commands, with optional dry-run and verbose modes.
type CmdRunner struct {
	DryRun  bool
	Verbose bool
	// Out receives dry-run and verbose log lines; defaults to os.Stderr when nil.
	Out io.Writer
}

// New constructs a CmdRunner.
func New(dryRun, verbose bool) *CmdRunner {
	return &CmdRunner{DryRun: dryRun, Verbose: verbose}
}

// Run executes name with args, returning captured stdout and stderr.
func (r *CmdRunner) Run(name string, args ...string) (string, string, error) {
	return r.RunDir("", nil, name, args...)
}

// RunEnv executes name with args, appending env to the current process environment.
func (r *CmdRunner) RunEnv(env []string, name string, args ...string) (string, string, error) {
	return r.RunDir("", env, name, args...)
}

// RunDir executes name with args in dir (empty = current dir), appending env.
func (r *CmdRunner) RunDir(dir string, env []string, name string, args ...string) (string, string, error) {
	if r.DryRun {
		_, _ = fmt.Fprintf(r.writer(), "[dry-run] %s %s\n", name, strings.Join(args, " "))
		return "", "", nil
	}

	if r.Verbose {
		_, _ = fmt.Fprintf(r.writer(), "[exec] %s %s\n", name, strings.Join(args, " "))
	}

	cmd := os_exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	if r.Verbose {
		echoOutput(r.writer(), stdout.String(), stderr.String())
	}

	if runErr != nil {
		if se := strings.TrimSpace(stderr.String()); se != "" {
			return stdout.String(), stderr.String(), fmt.Errorf("%s: %w: %s", name, runErr, se)
		}
		return stdout.String(), stderr.String(), fmt.Errorf("%s: %w", name, runErr)
	}

	return stdout.String(), stderr.String(), nil
}

func (r *CmdRunner) writer() io.Writer {
	if r.Out != nil {
		return r.Out
	}
	return os.Stderr
}

// echoOutput writes a command's captured stdout and stderr to w, each non-empty
// line indented by two spaces, so verbose runs show what each command produced.
func echoOutput(w io.Writer, stdout, stderr string) {
	for _, block := range []string{stdout, stderr} {
		for _, line := range strings.Split(strings.TrimRight(block, "\n"), "\n") {
			if line == "" {
				continue
			}
			_, _ = fmt.Fprintf(w, "  %s\n", line)
		}
	}
}
