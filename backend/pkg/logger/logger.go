package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Config defines structured logging options.
type Config struct {
	Environment string
	Level       string
	AddSource   bool
	Output      io.Writer
}

// ParseLevel parses a string representation of slog.Level.
func ParseLevel(lvl string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(lvl)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// New creates a configured slog.Logger with environment-appropriate formatting.
// In production, it emits structured JSON. In development/test, it emits human-readable text.
func New(cfg Config) *slog.Logger {
	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	level := ParseLevel(cfg.Level)
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if strings.ToLower(cfg.Environment) == "production" {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	return slog.New(handler)
}

// NewDefault returns a standard logger configured for local development.
func NewDefault() *slog.Logger {
	return New(Config{
		Environment: "development",
		Level:       "info",
		AddSource:   false,
	})
}
