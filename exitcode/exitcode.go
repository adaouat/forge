// Package exitcode maps errors to process exit codes. Lower layers return plain
// or sentinel errors; the cmd boundary annotates them with a code via Wrap (or
// returns an ExitError directly), and main resolves the final code via Resolve.
// This package is a leaf and defines no code values — those are app policy.
package exitcode

import "errors"

// ExitError carries a process exit code alongside an error or message.
// Construct it directly with a Code and Message, or via Wrap to annotate an
// existing error.
type ExitError struct {
	Code    int
	Message string
	Err     error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *ExitError) Unwrap() error { return e.Err }

// Wrap annotates err with an exit code. It returns nil when err is nil. If err
// already carries an exit code (anywhere in its chain), the original code is
// preserved — the first/innermost classification wins.
func Wrap(code int, err error) error {
	if err == nil {
		return nil
	}
	var ee *ExitError
	if errors.As(err, &ee) {
		return err
	}
	return &ExitError{Code: code, Err: err}
}

// Resolve maps an error to an exit code: nil → 0, an error carrying an exit code
// → that code, anything else → 1 (the generic failure default).
func Resolve(err error) int {
	if err == nil {
		return 0
	}
	var ee *ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return 1
}
