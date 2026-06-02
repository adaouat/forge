// Package exectest provides test doubles for the exec package: a recording
// MockRunner and a FakeBin helper for tests that need the real exec path.
package exectest

import (
	"fmt"

	"github.com/adaouat/forge/exec"
)

var _ exec.Runner = (*MockRunner)(nil)

// Call records a single invocation of MockRunner.Run or MockRunner.RunEnv.
type Call struct {
	Name string
	Args []string
	Env  []string // nil for Run calls; set for RunEnv calls
}

type queuedResponse struct {
	Stdout string
	Stderr string
	Err    error
}

// MockRunner is an exec.Runner that records calls and returns canned responses.
type MockRunner struct {
	Calls     []Call
	responses []queuedResponse
}

// NewMockRunner constructs an empty MockRunner.
func NewMockRunner() *MockRunner {
	return &MockRunner{}
}

// QueueResponse enqueues one response to be returned by the next Run/RunEnv call.
func (m *MockRunner) QueueResponse(stdout, stderr string, err error) {
	m.responses = append(m.responses, queuedResponse{stdout, stderr, err})
}

// Run records the call and returns the next queued response.
func (m *MockRunner) Run(name string, args ...string) (string, string, error) {
	return m.RunEnv(nil, name, args...)
}

// RunEnv records the call (including env) and returns the next queued response.
func (m *MockRunner) RunEnv(env []string, name string, args ...string) (string, string, error) {
	m.Calls = append(m.Calls, Call{Name: name, Args: args, Env: env})
	if len(m.responses) == 0 {
		return "", "", fmt.Errorf("MockRunner: no response queued for %q", name)
	}
	r := m.responses[0]
	m.responses = m.responses[1:]
	return r.Stdout, r.Stderr, r.Err
}
