package ui

// Mode is a CLI's output mode. Spinners, progress bars, and colors are only
// shown in Human mode on a real TTY. The zero value is Human.
type Mode int

const (
	Human Mode = iota // styled, interactive output for a human at a TTY
	Plain             // unstyled plain text, for CI and pipes
	JSON              // newline-delimited JSON events
)

// ParseMode maps a flag value to a Mode, defaulting to Human for unknown input.
func ParseMode(s string) Mode {
	switch s {
	case "plain":
		return Plain
	case "json":
		return JSON
	default:
		return Human
	}
}

// IsHuman reports whether m is the human (styled) output mode.
func (m Mode) IsHuman() bool { return m == Human }

// String returns the mode's flag name.
func (m Mode) String() string {
	switch m {
	case Plain:
		return "plain"
	case JSON:
		return "json"
	default:
		return "human"
	}
}
