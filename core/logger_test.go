package core

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestStructuredLogger(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	testLog := slog.New(handler)

	testLog.Info("test_event", slog.String("component", "ipv7_core"), slog.Int("peers", 42))

	output := buf.String()
	if !strings.Contains(output, `"msg":"test_event"`) {
		t.Fatalf("expected json message 'test_event', got: %s", output)
	}
	if !strings.Contains(output, `"peers":42`) {
		t.Fatalf("expected json field 'peers:42', got: %s", output)
	}
}
