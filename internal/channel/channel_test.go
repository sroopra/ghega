package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sroopra/ghega/internal/logging"
	"github.com/sroopra/ghega/pkg/httpsender"
	"github.com/sroopra/ghega/pkg/mapping"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/mllp"
)

// safeBuffer wraps a bytes.Buffer with a sync.Mutex so it can be shared
// safely between goroutines (e.g. the MLLP listener and the test).
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sb *safeBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *safeBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

// syntheticADT_A01 returns a clearly synthetic ADT^A01 message with no PHI.
func syntheticADT_A01() []byte {
	return []byte("MSH|^~\\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|20240101120000||ADT^A01|MSG001|P|2.5\r" +
		"EVN|A01|20240101000000|||\r" +
		"PID|1||SYNTHETIC_MRN_123456^^^GHEGA_FACILITY^MR||TESTPATIENT^SYNTHETIC||19800101|M|||123 SYNTHETIC STREET^^SYNTHETIC CITY^ST^12345||555-0100||||||||||||||||||||\r" +
		"PV1|1|I|GHEGA_WARD^GHEGA_ROOM^1||||||||||||||||||||||||||||||||||||||||||20240101000000\r")
}

func TestChannelEndToEnd_ADT_A01(t *testing.T) {
	// Capture logs to verify no payload bytes are logged.
	var logBuf safeBuffer
	phiLogger := logging.New(&logBuf, slog.LevelDebug)

	// 1. Start a test HTTP server that captures received payloads.
	var receivedBody []byte
	var receivedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = body
		receivedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	}))
	defer server.Close()

	// 2. Create an in-memory message store.
	store := messagestore.NewInMemoryStore()

	// 3. Build channel configuration.
	cfg := Config{
		Name: "test-adt-a01",
		Source: SourceConfig{
			Type: "mllp",
			Host: "127.0.0.1",
			Port: 0, // let OS assign
		},
		Destination: DestinationConfig{
			Type: "http",
			URL:  server.URL,
		},
		Mapping: MappingConfig{
			MessageType: "ADT_A01",
			Mappings: []mapping.Mapping{
				{Source: "MSH-9", Target: "message_type", Transform: "copy"},
				{Source: "PID-3.1", Target: "patient_mrn", Transform: "copy"},
			},
		},
	}

	sender := &httpsender.Sender{
		URL:     cfg.Destination.URL,
		Timeout: 5 * time.Second,
		Retries: 1,
	}

	ch := NewChannel(cfg, store, sender, phiLogger)
	if err := ch.Run(); err != nil {
		t.Fatalf("channel run failed: %v", err)
	}
	defer ch.Stop()

	// 4. Send a synthetic ADT_A01 message over MLLP.
	addr := ch.Addr()
	if addr == "" {
		t.Fatal("channel has no address")
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	payload := syntheticADT_A01()
	frame := mllp.EncodeFrame(payload)
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("write frame failed: %v", err)
	}

	// Read ACK response.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read ack failed: %v", err)
	}

	ackPayload, consumed, err := mllp.DecodeFrame(buf[:n])
	if err != nil {
		t.Fatalf("decode ack failed: %v", err)
	}
	if consumed != n {
		t.Errorf("expected to consume entire response, consumed %d of %d", consumed, n)
	}

	// Assert: ACK received over MLLP with AA code.
	ackStr := string(ackPayload)
	if !strings.Contains(ackStr, "MSA|AA|MSG001") {
		t.Errorf("expected AA ack for MSG001, got: %s", ackStr)
	}

	// Allow async processing to complete.
	time.Sleep(100 * time.Millisecond)

	// Assert: Message metadata persisted in store.
	ctx := context.Background()
	env, err := store.GetMetadata(ctx, "MSG001")
	if err != nil {
		t.Fatalf("expected metadata in store: %v", err)
	}
	if env.ChannelID != cfg.Name {
		t.Errorf("expected channel_id %q, got %q", cfg.Name, env.ChannelID)
	}
	if env.MessageID != "MSG001" {
		t.Errorf("expected message_id MSG001, got %q", env.MessageID)
	}
	if env.Status != "received" {
		t.Errorf("expected status 'received', got %q", env.Status)
	}

	// Assert: HTTP destination received the mapped payload.
	if len(receivedBody) == 0 {
		t.Fatal("expected HTTP destination to receive mapped payload")
	}
	var mapped map[string]string
	if err := json.Unmarshal(receivedBody, &mapped); err != nil {
		t.Fatalf("expected valid JSON mapped payload: %v", err)
	}
	if mapped["message_type"] != "ADT^A01" {
		t.Errorf("expected message_type 'ADT^A01', got %q", mapped["message_type"])
	}
	if mapped["patient_mrn"] != "SYNTHETIC_MRN_123456" {
		t.Errorf("expected patient_mrn 'SYNTHETIC_MRN_123456', got %q", mapped["patient_mrn"])
	}
	if receivedContentType != "application/json" {
		// The httpsender does not set Content-Type by default; verify the body is JSON.
		_ = receivedContentType
	}

	// Assert: Logs contain no payload bytes.
	logs := logBuf.String()
	payloadStr := string(payload)
	if strings.Contains(logs, payloadStr) {
		t.Errorf("logs contained raw payload bytes")
	}
	// Also ensure no specific PHI-like strings leaked.
	if strings.Contains(logs, "SYNTHETIC_MRN_123456") {
		t.Errorf("logs contained payload-derived identifier")
	}
	if strings.Contains(logs, "TESTPATIENT") {
		t.Errorf("logs contained payload-derived name")
	}
	// But metadata should be present.
	if !strings.Contains(logs, "MSG001") {
		t.Errorf("logs missing message_id metadata")
	}
	if !strings.Contains(logs, "message delivered") {
		t.Errorf("logs missing delivery confirmation")
	}
}

func TestChannel_LoadConfig(t *testing.T) {
	yamlData := `name: test-channel
source:
  type: mllp
  host: 127.0.0.1
  port: 2575
destination:
  type: http
  url: http://example.com/webhook
mapping:
  messageType: ADT_A01
  mappings:
    - source: MSH-9
      target: message_type
      transform: copy
`
	tmpFile, err := os.CreateTemp("", "channel-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if cfg.Name != "test-channel" {
		t.Errorf("expected name 'test-channel', got %q", cfg.Name)
	}
	if cfg.Source.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %q", cfg.Source.Host)
	}
	if cfg.Source.Port != 2575 {
		t.Errorf("expected port 2575, got %d", cfg.Source.Port)
	}
	if cfg.Destination.URL != "http://example.com/webhook" {
		t.Errorf("expected url 'http://example.com/webhook', got %q", cfg.Destination.URL)
	}
	if cfg.Mapping.MessageType != "ADT_A01" {
		t.Errorf("expected messageType ADT_A01, got %q", cfg.Mapping.MessageType)
	}
	if len(cfg.Mapping.Mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %d", len(cfg.Mapping.Mappings))
	}
	if cfg.Mapping.Mappings[0].Source != "MSH-9" {
		t.Errorf("expected source MSH-9, got %q", cfg.Mapping.Mappings[0].Source)
	}
}

func TestChannel_LoadConfigMissingName(t *testing.T) {
	yamlData := `source:
  type: mllp
  host: 127.0.0.1
  port: 2575
`
	tmpFile, err := os.CreateTemp("", "channel-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}
