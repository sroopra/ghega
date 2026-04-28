package engine

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sroopra/ghega/pkg/hl7v2"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/mllp"
)

func TestMLLPHandler_EndToEnd(t *testing.T) {
	store := messagestore.NewInMemoryStore()

	// Create a test HTTP destination server.
	destSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer destSrv.Close()

	cfg := DefaultHandlerConfig()
	cfg.DestinationURL = destSrv.URL
	cfg.Timeout = 2 * time.Second

	handler := NewMLLPHandler(store, cfg, nil)

	// Build a synthetic ADT A01 message.
	msg := &hl7v2.Message{
		Segments: []hl7v2.Segment{
			{Type: "MSH", Fields: []string{"|", "^~\\&", "GHEGA", "FACILITY", "RECEIVER", "FACILITY", time.Now().Format("20060102150405"), "", "ADT^A01", "MSG001", "P", "2.5"}},
			{Type: "EVN", Fields: []string{"A01", "20240101000000"}},
			{Type: "PID", Fields: []string{"1", "", "SYNTHETIC_MRN_123456^^^FACILITY^MR", "", "TESTPATIENT^SYNTHETIC", "", "19800101", "M"}},
		},
		FieldSeparator:  '|',
		ComponentSep:    '^',
		RepetitionSep:   '~',
		EscapeChar:      '\\',
		SubcomponentSep: '&',
	}

	ackPayload, err := handler(msg)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// Verify ACK was returned.
	ack, err := hl7v2.Parse(ackPayload)
	if err != nil {
		t.Fatalf("parse ack: %v", err)
	}
	if ack.Segment("MSA") == nil {
		t.Fatal("ack missing MSA segment")
	}
	if ack.Segment("MSA").Field(1) != "AA" {
		t.Errorf("expected AA ack, got %s", ack.Segment("MSA").Field(1))
	}

	// Verify message was persisted.
	ctx := context.Background()
	envelopes, err := store.ListAll(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(envelopes) != 1 {
		t.Fatalf("expected 1 message, got %d", len(envelopes))
	}

	env := envelopes[0]
	if env.ChannelID != "adt-a01" {
		t.Errorf("channel_id = %q, want adt-a01", env.ChannelID)
	}
	if env.Status != "delivered" {
		t.Errorf("status = %q, want delivered", env.Status)
	}

	// Verify payload was stored.
	payload, ok := store.GetPayload(env.Ref.StorageID)
	if !ok {
		t.Fatal("payload not found")
	}
	if len(payload) == 0 {
		t.Error("payload is empty")
	}
}

func TestMLLPHandler_DestinationFailure(t *testing.T) {
	store := messagestore.NewInMemoryStore()

	// Destination that always fails.
	destSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer destSrv.Close()

	cfg := DefaultHandlerConfig()
	cfg.DestinationURL = destSrv.URL
	cfg.Timeout = 2 * time.Second
	cfg.Retries = 1

	handler := NewMLLPHandler(store, cfg, nil)

	msg := &hl7v2.Message{
		Segments: []hl7v2.Segment{
			{Type: "MSH", Fields: []string{"|", "^~\\&", "GHEGA", "FACILITY", "RECEIVER", "FACILITY", time.Now().Format("20060102150405"), "", "ADT^A01", "MSG002", "P", "2.5"}},
			{Type: "PID", Fields: []string{"1", "", "SYNTHETIC_MRN_123456^^^FACILITY^MR", "", "TESTPATIENT^SYNTHETIC", "", "19800101", "M"}},
		},
		FieldSeparator:  '|',
		ComponentSep:    '^',
		RepetitionSep:   '~',
		EscapeChar:      '\\',
		SubcomponentSep: '&',
	}

	ackPayload, err := handler(msg)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// Still returns AA ACK.
	ack, _ := hl7v2.Parse(ackPayload)
	if ack.Segment("MSA").Field(1) != "AA" {
		t.Errorf("expected AA ack even on failure, got %s", ack.Segment("MSA").Field(1))
	}

	// Message status should be failed.
	ctx := context.Background()
	envelopes, _ := store.ListAll(ctx, 10, 0)
	if len(envelopes) != 1 {
		t.Fatalf("expected 1 message, got %d", len(envelopes))
	}
	if envelopes[0].Status != "failed" {
		t.Errorf("status = %q, want failed", envelopes[0].Status)
	}
}

func TestMLLPHandler_IntegrationWithMLLPListener(t *testing.T) {
	store := messagestore.NewInMemoryStore()

	destSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer destSrv.Close()

	cfg := DefaultHandlerConfig()
	cfg.DestinationURL = destSrv.URL
	cfg.Timeout = 2 * time.Second

	listener := mllp.NewListener(mllp.Config{Host: "127.0.0.1", Port: 0}, NewMLLPHandler(store, cfg, nil), nil)
	if err := listener.Start(); err != nil {
		t.Fatalf("start listener: %v", err)
	}
	defer listener.Stop()

	addr := listener.Addr()
	if addr == nil {
		t.Fatal("listener addr is nil")
	}

	// Connect and send a framed HL7 message.
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		t.Fatalf("dial listener: %v", err)
	}
	defer conn.Close()

	hl7Payload := []byte("MSH|^~\\&|GHEGA|FACILITY|RECEIVER|FACILITY|20240101000000||ADT^A01|MSG003|P|2.5\rPID|1||SYNTHETIC_MRN_123456^^^FACILITY^MR||TESTPATIENT^SYNTHETIC||19800101|M\r")
	frame := mllp.EncodeFrame(hl7Payload)
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("write frame: %v", err)
	}

	// Read ACK.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read ack: %v", err)
	}

	ackPayload, _, err := mllp.DecodeFrame(buf[:n])
	if err != nil {
		t.Fatalf("decode ack: %v", err)
	}

	ack, err := hl7v2.Parse(ackPayload)
	if err != nil {
		t.Fatalf("parse ack: %v", err)
	}
	if ack.Segment("MSA").Field(1) != "AA" {
		t.Errorf("expected AA ack, got %s", ack.Segment("MSA").Field(1))
	}

	// Verify message persisted.
	ctx := context.Background()
	envelopes, _ := store.ListAll(ctx, 10, 0)
	if len(envelopes) != 1 {
		t.Fatalf("expected 1 message, got %d", len(envelopes))
	}
	if envelopes[0].Status != "delivered" {
		t.Errorf("status = %q, want delivered", envelopes[0].Status)
	}
}
