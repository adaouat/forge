package log_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adaouat/forge/log"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		level      slog.Level
		emit       func(*slog.Logger)
		wantText   string
		wantAbsent string
	}{
		{
			name:  "info level reports info, filters debug",
			level: slog.LevelInfo,
			emit: func(l *slog.Logger) {
				l.Info("info message")
				l.Debug("debug message")
			},
			wantText:   "info message",
			wantAbsent: "debug message",
		},
		{
			name:  "warn level reports warn, filters info",
			level: slog.LevelWarn,
			emit: func(l *slog.Logger) {
				l.Warn("warn message")
				l.Info("info message")
			},
			wantText:   "warn message",
			wantAbsent: "info message",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			logger := log.New(&buf, tc.level)
			tc.emit(logger)

			out := buf.String()
			assert.Contains(t, out, tc.wantText)
			assert.NotContains(t, out, tc.wantAbsent)
		})
	}
}

func TestNew_writesToProvidedWriter(t *testing.T) {
	var buf bytes.Buffer

	logger := log.New(&buf, slog.LevelInfo)
	logger.Info("routed message")

	assert.Contains(t, buf.String(), "routed message")
}
