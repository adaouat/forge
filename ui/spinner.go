package ui

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
)

var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)

// Result is a task's successful outcome: a Detail shown after the task name
// ("✓ name — Detail") and optional indented sub-result lines.
type Result struct {
	Detail string
	Subs   []string
}

type skipError struct{ detail string }

func (e *skipError) Error() string { return e.detail }

// Skip marks a task as advisory: Spinner.Run renders "! name — detail" instead
// of success or failure and returns nil.
func Skip(detail string) error { return &skipError{detail} }

// Spinner runs named tasks, animating a spinner while each runs (human mode on a
// TTY) and resolving it to a ✓/✗/! status line written to out.
type Spinner struct {
	out   io.Writer
	mode  Mode
	total int
	n     int
}

// NewSpinner constructs a Spinner writing to out, animating only in m's human mode.
func NewSpinner(out io.Writer, m Mode) *Spinner {
	return &Spinner{out: out, mode: m}
}

// Total enables an [N/total] counter prefix on each task line.
func (s *Spinner) Total(total int) *Spinner {
	s.total = total
	return s
}

// Run animates a spinner titled name while fn runs, then renders the outcome:
// (Result, nil) → success, Skip(detail) → advisory, any other error → failure
// (returned to the caller).
// nextLabel advances the step counter and returns the task label, prefixed
// with [N/total] when a Total is set.
func (s *Spinner) nextLabel(name string) string {
	s.n++
	if s.total > 0 {
		return fmt.Sprintf("[%d/%d] %s", s.n, s.total, name)
	}
	return name
}

// Step renders a numbered "✓ [N/total] name — detail" line for a step whose
// work already completed successfully, advancing the counter. Use it for steps
// whose work ran outside the spinner; use Run when the work should animate.
func (s *Spinner) Step(name, detail string) {
	line := s.nextLabel(name)
	if detail != "" {
		line += " — " + detail
	}
	_, _ = fmt.Fprintln(s.out, Success(s.out, line))
}

func (s *Spinner) Run(name string, fn func() (Result, error)) error {
	label := s.nextLabel(name)

	var (
		mu   sync.Mutex
		done bool
		stop chan struct{}
	)
	if s.mode.IsHuman() && IsTTY(s.out) {
		stop = make(chan struct{})
		go s.animate(label, &mu, &done, stop)
	}

	res, runErr := fn()

	if stop != nil {
		mu.Lock()
		done = true
		close(stop)
		// Clear the spinner line before the final result lands on a clean line.
		_, _ = fmt.Fprintf(s.out, "\r%s\r", strings.Repeat(" ", len(label)+10))
		mu.Unlock()
	}

	return s.render(label, res, runErr)
}

func (s *Spinner) animate(label string, mu *sync.Mutex, done *bool, stop <-chan struct{}) {
	sp := spinner.MiniDot
	ticker := time.NewTicker(sp.FPS)
	defer ticker.Stop()
	for i := 0; ; i++ {
		select {
		case <-stop:
			return
		case <-ticker.C:
			mu.Lock()
			if !*done {
				_, _ = fmt.Fprintf(s.out, "\r  %s  %s", spinnerStyle.Render(sp.Frames[i%len(sp.Frames)]), label)
			}
			mu.Unlock()
		}
	}
}

func (s *Spinner) render(label string, res Result, runErr error) error {
	var skip *skipError
	switch {
	case errors.As(runErr, &skip):
		_, _ = fmt.Fprintln(s.out, Warn(s.out, label+" — "+skip.detail))
		return nil
	case runErr != nil:
		lines := strings.Split(strings.TrimRight(runErr.Error(), "\n"), "\n")
		_, _ = fmt.Fprintln(s.out, Err(s.out, label+" — "+lines[0]))
		for _, line := range lines[1:] {
			_, _ = fmt.Fprintf(s.out, "       %s\n", strings.TrimLeft(line, " "))
		}
		return runErr
	default:
		line := label
		if res.Detail != "" {
			line = label + " — " + res.Detail
		}
		_, _ = fmt.Fprintln(s.out, Success(s.out, line))
		for _, sub := range res.Subs {
			_, _ = fmt.Fprintln(s.out, "     "+Success(s.out, sub))
		}
		return nil
	}
}
