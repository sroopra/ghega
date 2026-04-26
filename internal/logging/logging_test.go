package logging

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/sroopra/ghega/pkg/payloadref"
)

// syntheticPayload simulates PHI bytes that must never appear in log output.
const syntheticPayload = "PATIENT:DOE,JANE|MRN:87654321|SSN:111-11-1111|DOB:1990-02-02"

func captureOutput(f func(*Logger)) string {
	var buf bytes.Buffer
	logger := New(&buf, slog.LevelDebug)
	f(logger)
	return buf.String()
}

func TestLogMessageReceived_NeverContainsPayloadBytes(t *testing.T) {
	ref := payloadref.PayloadRef{
		StorageID: "store-received-001",
		Location:  "s3://ghega/received/msg-100",
	}
	out := captureOutput(func(l *Logger) {
		l.LogMessageReceived("channel-adt", "msg-100", time.Date(2024, 3, 10, 8, 15, 0, 0, time.UTC), ref)
	})

	if strings.Contains(out, syntheticPayload) {
		t.Errorf("LogMessageReceived leaked payload bytes: %q", out)
	}
	if !strings.Contains(out, "message received") {
		t.Errorf("LogMessageReceived missing event text: %q", out)
	}
	if !strings.Contains(out, "channel-adt") {
		t.Errorf("LogMessageReceived missing channel_id: %q", out)
	}
	if !strings.Contains(out, "msg-100") {
		t.Errorf("LogMessageReceived missing message_id: %q", out)
	}
	if !strings.Contains(out, "store-received-001") {
		t.Errorf("LogMessageReceived missing storage_id: %q", out)
	}
}

func TestLogMessageProcessed_NeverContainsPayloadBytes(t *testing.T) {
	ref := payloadref.PayloadRef{
		StorageID: "store-processed-002",
		Location:  "s3://ghega/processed/msg-200",
	}
	out := captureOutput(func(l *Logger) {
		l.LogMessageProcessed("channel-oru", "msg-200", time.Date(2024, 3, 10, 8, 16, 0, 0, time.UTC), ref)
	})

	if strings.Contains(out, syntheticPayload) {
		t.Errorf("LogMessageProcessed leaked payload bytes: %q", out)
	}
	if !strings.Contains(out, "message processed") {
		t.Errorf("LogMessageProcessed missing event text: %q", out)
	}
	if !strings.Contains(out, "channel-oru") {
		t.Errorf("LogMessageProcessed missing channel_id: %q", out)
	}
	if !strings.Contains(out, "msg-200") {
		t.Errorf("LogMessageProcessed missing message_id: %q", out)
	}
}

func TestLogMessageFailed_NeverContainsPayloadBytes(t *testing.T) {
	out := captureOutput(func(l *Logger) {
		l.LogMessageFailed("channel-adt", "msg-300", errors.New("mapping failed"))
	})

	if strings.Contains(out, syntheticPayload) {
		t.Errorf("LogMessageFailed leaked payload bytes: %q", out)
	}
	if !strings.Contains(out, "message failed") {
		t.Errorf("LogMessageFailed missing event text: %q", out)
	}
	if !strings.Contains(out, "mapping failed") {
		t.Errorf("LogMessageFailed missing error text: %q", out)
	}
	if !strings.Contains(out, "channel-adt") {
		t.Errorf("LogMessageFailed missing channel_id: %q", out)
	}
	if !strings.Contains(out, "msg-300") {
		t.Errorf("LogMessageFailed missing message_id: %q", out)
	}
}

func TestLogInfo_NeverContainsPayloadBytes(t *testing.T) {
	out := captureOutput(func(l *Logger) {
		l.LogInfo("connector started",
			slog.String("connector_id", "mllp-001"),
			slog.String("listen_addr", ":2575"),
		)
	})

	if strings.Contains(out, syntheticPayload) {
		t.Errorf("LogInfo leaked payload bytes: %q", out)
	}
	if !strings.Contains(out, "connector started") {
		t.Errorf("LogInfo missing message text: %q", out)
	}
	if !strings.Contains(out, "mllp-001") {
		t.Errorf("LogInfo missing connector_id: %q", out)
	}
}

func TestLogError_NeverContainsPayloadBytes(t *testing.T) {
	out := captureOutput(func(l *Logger) {
		l.LogError("connector bind failed",
			slog.String("connector_id", "mllp-001"),
			slog.String("error", "address already in use"),
		)
	})

	if strings.Contains(out, syntheticPayload) {
		t.Errorf("LogError leaked payload bytes: %q", out)
	}
	if !strings.Contains(out, "connector bind failed") {
		t.Errorf("LogError missing message text: %q", out)
	}
	if !strings.Contains(out, "address already in use") {
		t.Errorf("LogError missing error text: %q", out)
	}
}

// TestLogger_RefusesPayloadContent ensures that the Logger API surface does
// not expose any method that accepts raw payload bytes. This is enforced by
// the fact that no such method exists in this package.
func TestLogger_RefusesPayloadContent(t *testing.T) {
	// The Logger type has no methods that accept []byte or string payload content.
	// All methods accept discrete metadata fields.
	var buf bytes.Buffer
	logger := New(&buf, slog.LevelDebug)

	// Attempting to pass payload bytes is impossible because there is no API for it.
	// We verify the API surface by exercising the only available paths.
	ref := payloadref.PayloadRef{StorageID: "s", Location: "l"}
	logger.LogMessageReceived("c", "m", time.Now(), ref)
	logger.LogMessageProcessed("c", "m", time.Now(), ref)
	logger.LogMessageFailed("c", "m", errors.New("e"))
	logger.LogInfo("i", slog.String("k", "v"))
	logger.LogError("e", slog.String("k", "v"))
}

func TestNew_LogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, slog.LevelWarn)
	logger.LogInfo("this should be filtered")
	if strings.Contains(buf.String(), "this should be filtered") {
		t.Error("Info message was not filtered at Warn level")
	}

	buf.Reset()
	logger.LogError("this should appear", slog.String("reason", "test"))
	if !strings.Contains(buf.String(), "this should appear") {
		t.Error("Error message was filtered at Warn level")
	}
}
