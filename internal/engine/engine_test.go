package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sroopra/ghega/pkg/hl7v2"
	"github.com/sroopra/ghega/pkg/messagestore"
)

func TestEngine_HandleMessage_Delivered(t *testing.T) {
	// Start a test destination server.
	var received []byte
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = body
		w.WriteHeader(http.StatusOK)
	})
	srv := &http.Server{Handler: mux}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()

	store := messagestore.NewInMemoryStore()
	eng := NewEngine(store, nil)
	eng.Sender.URL = fmt.Sprintf("http://%s/webhook", ln.Addr().String())
	eng.Sender.Retries = 1

	msg := buildTestMessage()
	ack, err := eng.HandleMessage(msg)
	if err != nil {
		t.Fatalf("handle message error: %v", err)
	}
	if !strings.Contains(string(ack), "MSA|AA|") {
		t.Errorf("expected AA ack, got: %s", string(ack))
	}

	// Allow async sender goroutine to complete.
	time.Sleep(200 * time.Millisecond)

	if received == nil {
		t.Fatal("destination did not receive payload")
	}
	var mapped map[string]string
	if err := json.Unmarshal(received, &mapped); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if mapped["patient_mrn"] != "SYNTHETIC_MRN" {
		t.Errorf("expected patient_mrn=SYNTHETIC_MRN, got %s", mapped["patient_mrn"])
	}
	if mapped["patient_last_name"] != "TESTPATIENT" {
		t.Errorf("expected patient_last_name=TESTPATIENT, got %s", mapped["patient_last_name"])
	}
	if mapped["patient_first_name"] != "JOHN" {
		t.Errorf("expected patient_first_name=JOHN, got %s", mapped["patient_first_name"])
	}
	if mapped["message_type"] != "ADT" {
		t.Errorf("expected message_type=ADT, got %s", mapped["message_type"])
	}

	// Verify store state.
	ctx := t.Context()
	envelopes, err := store.ListByChannel(ctx, "ADT-A01", 10, 0)
	if err != nil {
		t.Fatalf("list by channel error: %v", err)
	}
	if len(envelopes) != 1 {
		t.Fatalf("expected 1 envelope, got %d", len(envelopes))
	}
	if envelopes[0].Status != "delivered" {
		t.Errorf("expected status delivered, got %s", envelopes[0].Status)
	}
}

func TestEngine_HandleMessage_Failed(t *testing.T) {
	// Start a test destination server that always fails.
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := &http.Server{Handler: mux}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()

	store := messagestore.NewInMemoryStore()
	eng := NewEngine(store, nil)
	eng.Sender.URL = fmt.Sprintf("http://%s/webhook", ln.Addr().String())
	eng.Sender.Retries = 1

	msg := buildTestMessage()
	ack, err := eng.HandleMessage(msg)
	if err != nil {
		t.Fatalf("handle message error: %v", err)
	}
	if !strings.Contains(string(ack), "MSA|AA|") {
		t.Errorf("expected AA ack even on failure, got: %s", string(ack))
	}

	time.Sleep(200 * time.Millisecond)

	ctx := t.Context()
	envelopes, err := store.ListByChannel(ctx, "ADT-A01", 10, 0)
	if err != nil {
		t.Fatalf("list by channel error: %v", err)
	}
	if len(envelopes) != 1 {
		t.Fatalf("expected 1 envelope, got %d", len(envelopes))
	}
	if envelopes[0].Status != "failed" {
		t.Errorf("expected status failed, got %s", envelopes[0].Status)
	}
}

func buildTestMessage() *hl7v2.Message {
	raw := "MSH|^~\\&|SENDING_APP|SENDING_FACILITY|RECEIVING_APP|RECEIVING_FACILITY|20240101120000||ADT^A01|MSG001|P|2.5\r" +
		"PID|1||SYNTHETIC_MRN^^^MRN||TESTPATIENT^JOHN^MICHAEL||19800101|M|||123 MAIN ST^^ANYTOWN^ST^12345||555-555-5555|||||SINGLE\r"
	msg, err := hl7v2.Parse([]byte(raw))
	if err != nil {
		panic(err)
	}
	return msg
}
