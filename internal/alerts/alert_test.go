package alerts

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestInMemoryAlertStore_CRUD(t *testing.T) {
	s := NewInMemoryAlertStore()

	// Create
	a := &Alert{
		ID:        "alert-001",
		ChannelID: "ch-1",
		Severity:  SeverityWarning,
		Message:   "Test alert message",
		CreatedAt: time.Now().UTC(),
	}
	if err := s.Create(a); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// List
	list, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(list))
	}
	if list[0].ID != a.ID {
		t.Errorf("expected alert ID %q, got %q", a.ID, list[0].ID)
	}

	// Acknowledge
	if err := s.Acknowledge(a.ID); err != nil {
		t.Fatalf("Acknowledge failed: %v", err)
	}
	list, _ = s.List()
	if list[0].AcknowledgedAt == nil {
		t.Error("expected AcknowledgedAt to be set")
	}

	// Resolve
	if err := s.Resolve(a.ID); err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	list, _ = s.List()
	if list[0].ResolvedAt == nil {
		t.Error("expected ResolvedAt to be set")
	}

	// Acknowledge unknown
	if err := s.Acknowledge("unknown"); err == nil {
		t.Error("expected error for unknown alert ID")
	}

	// Resolve unknown
	if err := s.Resolve("unknown"); err == nil {
		t.Error("expected error for unknown alert ID")
	}
}

func TestTriggerLog(t *testing.T) {
	var captured string
	logFn := func(format string, args ...any) {
		captured = fmt.Sprintf(format, args...)
	}

	TriggerLog(logFn, "ch-1", "msg-1", errors.New("something failed"))
	if captured == "" {
		t.Error("expected TriggerLog to capture a message")
	}
	if !strings.Contains(captured, "ch-1") || !strings.Contains(captured, "msg-1") {
		t.Errorf("expected captured log to contain channel and message IDs, got: %s", captured)
	}
}
