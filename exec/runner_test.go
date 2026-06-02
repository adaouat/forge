package exec_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/forge/exec"
)

func TestCmdRunner_Run_success(t *testing.T) {
	r := exec.New(false, false)
	stdout, stderr, err := r.Run("sh", "-c", "echo hello")
	require.NoError(t, err)
	assert.Equal(t, "hello\n", stdout)
	assert.Empty(t, stderr)
}

func TestCmdRunner_Run_dryRun(t *testing.T) {
	// A nonexistent binary would fail if actually executed.
	r := exec.New(true, false)
	stdout, stderr, err := r.Run("nonexistent_xyzzy_forge_abc")
	require.NoError(t, err)
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
}

func TestCmdRunner_Run_failure(t *testing.T) {
	r := exec.New(false, false)
	stdout, stderr, err := r.Run("sh", "-c", "echo oops >&2; exit 1")
	require.Error(t, err)
	assert.Empty(t, stdout)
	assert.Equal(t, "oops\n", stderr)
}

func TestCmdRunner_RunEnv_propagatesEnv(t *testing.T) {
	r := exec.New(false, false)
	stdout, _, err := r.RunEnv([]string{"MY_TEST_VAR=testvalue"}, "sh", "-c", "echo $MY_TEST_VAR")
	require.NoError(t, err)
	assert.Equal(t, "testvalue\n", stdout)
}

func TestCmdRunner_RunEnv_dryRun(t *testing.T) {
	r := exec.New(true, false)
	stdout, stderr, err := r.RunEnv([]string{"KEY=val"}, "nonexistent_xyzzy_forge_abc")
	require.NoError(t, err)
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
}

func TestCmdRunner_Run_verbose(t *testing.T) {
	var buf bytes.Buffer
	r := exec.New(false, true)
	r.Out = &buf
	stdout, _, err := r.Run("sh", "-c", "echo done")
	require.NoError(t, err)
	assert.Equal(t, "done\n", stdout)
	assert.Contains(t, buf.String(), "[exec] sh")
}

func TestCmdRunner_Run_verbose_echoesOutput(t *testing.T) {
	var buf bytes.Buffer
	r := exec.New(false, true)
	r.Out = &buf
	_, _, err := r.Run("sh", "-c", "echo out-line; echo err-line >&2")
	require.NoError(t, err)

	logged := buf.String()
	assert.Contains(t, logged, "[exec] sh")
	assert.Contains(t, logged, "out-line", "verbose should echo captured stdout")
	assert.Contains(t, logged, "err-line", "verbose should echo captured stderr")
	// Echoed output lines are indented under the [exec] line.
	assert.Contains(t, logged, "  out-line")
	assert.Contains(t, logged, "  err-line")
}

func TestCmdRunner_Run_failure_includesStderr(t *testing.T) {
	r := exec.New(false, false)
	_, _, err := r.Run("sh", "-c", "echo 'boom detail' >&2; exit 1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sh", "error should name the failing command")
	assert.Contains(t, err.Error(), "boom detail", "error should carry the command's stderr")
}

func TestCmdRunner_Run_failure_noStderr_cleanMessage(t *testing.T) {
	r := exec.New(false, false)
	_, _, err := r.Run("sh", "-c", "exit 1")
	require.Error(t, err)
	// No stderr → no trailing ": " artifact appended after the exit status.
	assert.False(t, strings.HasSuffix(err.Error(), ": "), "error message should not end with a dangling colon")
	assert.Contains(t, err.Error(), "sh")
}

func TestCmdRunner_RunDir_runsInGivenDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "marker.txt"), []byte("in-dir"), 0o600))
	r := exec.New(false, false)
	// Reading a relative path only succeeds if the command runs inside dir.
	stdout, _, err := r.RunDir(dir, nil, "sh", "-c", "cat marker.txt")
	require.NoError(t, err)
	assert.Equal(t, "in-dir", stdout)
}

func TestCmdRunner_RunDir_propagatesEnv(t *testing.T) {
	r := exec.New(false, false)
	stdout, _, err := r.RunDir(t.TempDir(), []string{"MARKER_VAR=val"}, "sh", "-c", "echo $MARKER_VAR")
	require.NoError(t, err)
	assert.Equal(t, "val\n", stdout)
}

func TestCmdRunner_RunDir_dryRun(t *testing.T) {
	r := exec.New(true, false)
	stdout, stderr, err := r.RunDir(t.TempDir(), nil, "nonexistent_xyzzy_forge_abc")
	require.NoError(t, err)
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
}

func TestCmdRunner_implementsRunner(t *testing.T) {
	var _ exec.Runner = exec.New(false, false)
}
