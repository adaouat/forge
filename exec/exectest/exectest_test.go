package exectest_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/forge/exec"
	"github.com/adaouat/forge/exec/exectest"
)

func TestMockRunner_implementsRunner(t *testing.T) {
	var _ exec.Runner = exectest.NewMockRunner()
}

func TestMockRunner_recordsCallsAndReturnsResponsesFIFO(t *testing.T) {
	wantErr := errors.New("boom")
	mr := exectest.NewMockRunner()
	mr.QueueResponse("out1", "err1", nil)
	mr.QueueResponse("out2", "err2", wantErr)

	stdout, stderr, err := mr.Run("git", "tag", "v1.2.3")
	require.NoError(t, err)
	assert.Equal(t, "out1", stdout)
	assert.Equal(t, "err1", stderr)

	stdout, stderr, err = mr.RunEnv([]string{"K=V"}, "gh", "release", "create")
	require.ErrorIs(t, err, wantErr)
	assert.Equal(t, "out2", stdout)
	assert.Equal(t, "err2", stderr)

	require.Len(t, mr.Calls, 2)
	assert.Equal(t, "git", mr.Calls[0].Name)
	assert.Equal(t, []string{"tag", "v1.2.3"}, mr.Calls[0].Args)
	assert.Nil(t, mr.Calls[0].Env, "Run records nil env")

	assert.Equal(t, "gh", mr.Calls[1].Name)
	assert.Equal(t, []string{"release", "create"}, mr.Calls[1].Args)
	assert.Equal(t, []string{"K=V"}, mr.Calls[1].Env, "RunEnv records the env")
}

func TestMockRunner_noResponseQueued_returnsError(t *testing.T) {
	mr := exectest.NewMockRunner()
	_, _, err := mr.Run("git", "status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "git", "error should name the unanswered command")
}

func TestFakeBin_installsRunnableScriptOnPath(t *testing.T) {
	exectest.FakeBin(t, "forge_greet", "#!/bin/sh\necho hi")
	r := exec.New(false, false)
	stdout, _, err := r.Run("forge_greet")
	require.NoError(t, err)
	assert.Equal(t, "hi\n", stdout)
}
