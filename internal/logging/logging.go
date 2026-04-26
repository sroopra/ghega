// Package logging provides a PHI-safe logger wrapper.
//
// The logger refuses to accept payload bytes in any of its methods.
// All logging is metadata-only. Use the explicit methods such as
// LogMessageReceived and LogMessageProcessed instead of generic
// printf-style logging when working with Envelope or PayloadRef data.
package logging

import (
	"io"
	"log/slog"
	"time"

	"github.com/sroopra/ghega/pkg/payloadref"
)

// Logger is a PHI-safe wrapper around a structured logger.
// It never exposes methods that accept raw payload bytes.
type Logger struct {
	inner *slog.Logger
}

// New creates a Logger that writes to out at the given level.
func New(out io.Writer, level slog.Level) *Logger {
	handler := slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: level,
	})
	return &Logger{inner: slog.New(handler)}
}

// LogMessageReceived logs that a message was received.
// Only metadata and the payload reference are logged; payload bytes are not accepted.
func (l *Logger) LogMessageReceived(channelID, messageID string, receivedAt time.Time, ref payloadref.PayloadRef) {
	l.inner.Info("message received",
		slog.String("channel_id", channelID),
		slog.String("message_id", messageID),
		slog.Time("received_at", receivedAt),
		slog.String("storage_id", ref.StorageID),
		slog.String("location", ref.Location),
	)
}

// LogMessageProcessed logs that a message was successfully processed.
// Only metadata and the payload reference are logged; payload bytes are not accepted.
func (l *Logger) LogMessageProcessed(channelID, messageID string, processedAt time.Time, ref payloadref.PayloadRef) {
	l.inner.Info("message processed",
		slog.String("channel_id", channelID),
		slog.String("message_id", messageID),
		slog.Time("processed_at", processedAt),
		slog.String("storage_id", ref.StorageID),
		slog.String("location", ref.Location),
	)
}

// LogMessageFailed logs that a message could not be processed.
// Only metadata and the error are logged; payload bytes are not accepted.
func (l *Logger) LogMessageFailed(channelID, messageID string, err error) {
	l.inner.Error("message failed",
		slog.String("channel_id", channelID),
		slog.String("message_id", messageID),
		slog.String("error", err.Error()),
	)
}

// LogInfo logs a generic informational event with structured fields.
// Callers must ensure that no field value contains payload bytes.
func (l *Logger) LogInfo(msg string, attrs ...slog.Attr) {
	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	l.inner.Info(msg, args...)
}

// LogError logs a generic error event with structured fields.
// Callers must ensure that no field value contains payload bytes.
func (l *Logger) LogError(msg string, attrs ...slog.Attr) {
	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	l.inner.Error(msg, args...)
}

// Inner returns the underlying *slog.Logger for use by subsystems that
// need a standard logger. Callers must ensure no payload bytes are passed
// through the returned logger.
func (l *Logger) Inner() *slog.Logger {
	if l == nil {
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return l.inner
}
