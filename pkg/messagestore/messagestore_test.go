package messagestore

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/sroopra/ghega/internal/logging"
	"github.com/sroopra/ghega/pkg/payloadref"
)

// syntheticPayload simulates PHI bytes that must never appear in log output.
const syntheticPayload = "PATIENT:DOE,JANE|MRN:87654321|SSN:111-11-1111|DOB:1990-02-02"

func newTestEnvelope(channelID, messageID, status string) *payloadref.Envelope {
	return &payloadref.Envelope{
		ChannelID:  channelID,
		MessageID:  messageID,
		ReceivedAt: time.Date(2024, 3, 10, 8, 15, 0, 0, time.UTC),
		Status:     status,
		Ref: payloadref.PayloadRef{
			StorageID: "store-" + messageID,
			Location:  "mem://ghega/" + messageID,
		},
	}
}

func TestInMemoryStore_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	env := newTestEnvelope("ch-adt", "msg-001", "received")
	payload := []byte(syntheticPayload)

	if err := store.Save(ctx, env, payload); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := store.GetMetadata(ctx, "msg-001")
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}
	if got.MessageID != "msg-001" {
		t.Errorf("MessageID = %q, want %q", got.MessageID, "msg-001")
	}
	if got.ChannelID != "ch-adt" {
		t.Errorf("ChannelID = %q, want %q", got.ChannelID, "ch-adt")
	}
	if got.Status != "received" {
		t.Errorf("Status = %q, want %q", got.Status, "received")
	}
	if got.Ref.StorageID != "store-msg-001" {
		t.Errorf("Ref.StorageID = %q, want %q", got.Ref.StorageID, "store-msg-001")
	}

	// Verify raw payload is stored and retrievable
	p, ok := store.GetPayload("store-msg-001")
	if !ok {
		t.Fatal("GetPayload returned false")
	}
	if string(p) != syntheticPayload {
		t.Errorf("Payload mismatch: got %q, want %q", string(p), syntheticPayload)
	}
}

func TestInMemoryStore_GetMetadata_NotFound(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	_, err := store.GetMetadata(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing message")
	}
	if _, ok := err.(*ErrNotFound); !ok {
		t.Fatalf("expected *ErrNotFound, got %T", err)
	}
}

func TestInMemoryStore_ListByChannel(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	for i := 0; i < 5; i++ {
		env := newTestEnvelope("ch-oru", fmt.Sprintf("msg-%03d", i), "received")
		if err := store.Save(ctx, env, []byte(fmt.Sprintf("payload-%d", i))); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	// Save a message for a different channel
	env := newTestEnvelope("ch-adt", "msg-adt-001", "received")
	if err := store.Save(ctx, env, []byte("adt-payload")); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	results, err := store.ListByChannel(ctx, "ch-oru", 3, 0)
	if err != nil {
		t.Fatalf("ListByChannel failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("len(results) = %d, want %d", len(results), 3)
	}

	results, err = store.ListByChannel(ctx, "ch-oru", 10, 3)
	if err != nil {
		t.Fatalf("ListByChannel failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("len(results) = %d, want %d", len(results), 2)
	}

	results, err = store.ListByChannel(ctx, "ch-adt", 10, 0)
	if err != nil {
		t.Fatalf("ListByChannel failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len(results) = %d, want %d", len(results), 1)
	}
	if results[0].MessageID != "msg-adt-001" {
		t.Errorf("MessageID = %q, want %q", results[0].MessageID, "msg-adt-001")
	}
}

func TestInMemoryStore_ListByChannel_Empty(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	results, err := store.ListByChannel(ctx, "ch-empty", 10, 0)
	if err != nil {
		t.Fatalf("ListByChannel failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0", len(results))
	}
}

func TestSQLiteStore_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	env := newTestEnvelope("ch-adt", "msg-002", "received")
	payload := []byte(syntheticPayload)

	if err := store.Save(ctx, env, payload); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := store.GetMetadata(ctx, "msg-002")
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}
	if got.MessageID != "msg-002" {
		t.Errorf("MessageID = %q, want %q", got.MessageID, "msg-002")
	}
	if got.ChannelID != "ch-adt" {
		t.Errorf("ChannelID = %q, want %q", got.ChannelID, "ch-adt")
	}
	if got.Status != "received" {
		t.Errorf("Status = %q, want %q", got.Status, "received")
	}
	if got.Ref.StorageID != "store-msg-002" {
		t.Errorf("Ref.StorageID = %q, want %q", got.Ref.StorageID, "store-msg-002")
	}

	p, ok, err := store.GetPayload(ctx, "store-msg-002")
	if err != nil {
		t.Fatalf("GetPayload failed: %v", err)
	}
	if !ok {
		t.Fatal("GetPayload returned false")
	}
	if string(p) != syntheticPayload {
		t.Errorf("Payload mismatch: got %q, want %q", string(p), syntheticPayload)
	}
}

func TestSQLiteStore_GetMetadata_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	_, err = store.GetMetadata(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing message")
	}
	if _, ok := err.(*ErrNotFound); !ok {
		t.Fatalf("expected *ErrNotFound, got %T", err)
	}
}

func TestSQLiteStore_ListByChannel(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	for i := 0; i < 5; i++ {
		env := newTestEnvelope("ch-oru", fmt.Sprintf("msg-%03d", i), "received")
		if err := store.Save(ctx, env, []byte(fmt.Sprintf("payload-%d", i))); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	results, err := store.ListByChannel(ctx, "ch-oru", 3, 0)
	if err != nil {
		t.Fatalf("ListByChannel failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("len(results) = %d, want %d", len(results), 3)
	}

	results, err = store.ListByChannel(ctx, "ch-oru", 10, 3)
	if err != nil {
		t.Fatalf("ListByChannel failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("len(results) = %d, want %d", len(results), 2)
	}
}

func TestSQLiteStore_ListByChannel_Empty(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	results, err := store.ListByChannel(ctx, "ch-empty", 10, 0)
	if err != nil {
		t.Fatalf("ListByChannel failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0", len(results))
	}
}

func TestStore_NeverLogsPayloadBytes(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	var buf bytes.Buffer
	logger := logging.New(&buf, slog.LevelDebug)

	env := newTestEnvelope("ch-adt", "msg-log-test", "received")
	payload := []byte(syntheticPayload)

	if err := store.Save(ctx, env, payload); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Log the envelope metadata using the PHI-safe logger wrapper.
	logger.LogMessageReceived(env.ChannelID, env.MessageID, env.ReceivedAt, env.Ref)

	got, err := store.GetMetadata(ctx, "msg-log-test")
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}
	logger.LogMessageProcessed(got.ChannelID, got.MessageID, time.Now(), got.Ref)

	out := buf.String()
	if strings.Contains(out, syntheticPayload) {
		t.Errorf("log output leaked payload bytes: %q", out)
	}
	if !strings.Contains(out, "msg-log-test") {
		t.Errorf("log output missing message_id: %q", out)
	}
	if !strings.Contains(out, "store-msg-log-test") {
		t.Errorf("log output missing storage_id: %q", out)
	}
}

func TestEnvelope_String_NeverContainsPayloadBytes(t *testing.T) {
	env := newTestEnvelope("ch-adt", "msg-fmt", "received")
	env.Ref = payloadref.PayloadRef{StorageID: "sid", Location: "loc"}

	out := env.String()
	if strings.Contains(out, syntheticPayload) {
		t.Errorf("Envelope.String() leaked payload bytes: %q", out)
	}
}
