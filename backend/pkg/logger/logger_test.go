package logger_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/gitwise/backend/pkg/logger"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"invalid", slog.LevelInfo},
		{"", slog.LevelInfo},
	}

	for _, tc := range tests {
		got := logger.ParseLevel(tc.input)
		if got != tc.expected {
			t.Errorf("ParseLevel(%q) = %v, expected %v", tc.input, got, tc.expected)
		}
	}
}

func TestLogger_JSONProduction(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.Config{
		Environment: "production",
		Level:       "info",
		Output:      &buf,
	})

	l.Info("server started", slog.Int("port", 8080))

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("expected valid json from production logger: %v", err)
	}

	if m["msg"] != "server started" {
		t.Errorf("expected msg 'server started', got %v", m["msg"])
	}
	if m["level"] != "INFO" {
		t.Errorf("expected level 'INFO', got %v", m["level"])
	}
	if m["port"] != float64(8080) {
		t.Errorf("expected port 8080, got %v", m["port"])
	}
}

func TestLogger_TextDevelopment(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.Config{
		Environment: "development",
		Level:       "debug",
		Output:      &buf,
	})

	l.Debug("debug message", slog.String("key", "val"))

	out := buf.String()
	if !strings.Contains(out, "msg=\"debug message\"") && !strings.Contains(out, "debug message") {
		t.Errorf("expected text log to contain debug message, got: %s", out)
	}
	if !strings.Contains(out, "key=val") {
		t.Errorf("expected text log to contain key=val, got: %s", out)
	}
}

func TestNewDefault(t *testing.T) {
	l := logger.NewDefault()
	if l == nil {
		t.Fatal("expected non-nil default logger")
	}
}
