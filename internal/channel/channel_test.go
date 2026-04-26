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
	"testing"
	"time"

	"github.com/sroopra/ghega/internal/logging"
	"github.com/sroopra/ghega/pkg/mapping"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/mllp"
)

// syntheticADT_A01 is a clearly synthetic HL7v2 message with no real PHI.
const syntheticADT_A01 = "MSH|^~\\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|20240101120000||ADT^A01|MSG001|P|2.5\r" +
	"EVN|A01|20240101120000|\r" +
	"PID|1||SYNTHETIC_MRN_123456^^^GHEGA_FACILITY^MR||SYNTHETIC_PATIENT^TEST||19800101|M|||123 SYNTHETIC STREET^^SYNTHETIC CITY^ST^12345||555-0100||||||||||||||||||||\r"

func TestChannelEndToEnd(t *testing.T) {
	var logBuf bytes.Buffer
	logger := logging.New(&logBuf, slog.LevelDebug)

	// Destination HTTP server
	var receivedPayload []byte
	destServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedPayload = body
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer destServer.Close()

	cfg := &Config{
		Name: "test-adt-a01",
		Source: SourceConfig{
			Type: "mllp",
			Host: "127.0.0.1",
			Port: 0,
		},
		Destination: DestinationConfig{
			Type: "http",
			URL:  destServer.URL,
		},
		Mapping: MappingConfig{
			MessageType: "ADT_A01",
			Fields: []mapping.Mapping{
				{Source: "PID-3.1", Target: "patient_mrn", Transform: mapping.TransformCopy},
				{Source: "PID-5.1", Target: "last_name", Transform: mapping.TransformUppercase},
				{Target: "source_system", Transform: mapping.TransformStatic, Value: "ghega-test"},
			},
		},
	}

	store := messagestore.NewInMemoryStore()
	ch := NewChannel(cfg, store, logger)

	if err := ch.Run(); err != nil {
		t.Fatalf("channel run failed: %v", err)
	}
	defer ch.Stop()

	// Allow listener to bind
	time.Sleep(50 * time.Millisecond)

	addr := ch.Addr()
	if addr == "" {
		t.Fatal("channel has no address")
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	frame := mllp.EncodeFrame([]byte(syntheticADT_A01))
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Read ACK
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
	if !bytes.Contains(ackPayload, []byte("MSA|AA|MSG001")) {
		t.Errorf("expected AA ack for MSG001, got: %s", string(ackPayload))
	}

	// Wait for async HTTP send
	time.Sleep(100 * time.Millisecond)

	// Assert HTTP destination received mapped payload
	if len(receivedPayload) == 0 {
		t.Fatal("destination received no payload")
	}
	var mapped map[string]string
	if err := json.Unmarshal(receivedPayload, &mapped); err != nil {
		t.Fatalf("failed to unmarshal received payload: %v", err)
	}
	if mapped["patient_mrn"] != "SYNTHETIC_MRN_123456" {
		t.Errorf("patient_mrn = %q, want %q", mapped["patient_mrn"], "SYNTHETIC_MRN_123456")
	}
	if mapped["last_name"] != "SYNTHETIC_PATIENT" {
		t.Errorf("last_name = %q, want %q", mapped["last_name"], "SYNTHETIC_PATIENT")
	}
	if mapped["source_system"] != "ghega-test" {
		t.Errorf("source_system = %q, want %q", mapped["source_system"], "ghega-test")
	}

	// Assert metadata persisted in store
	var foundMeta bool
	list, err := store.ListByChannel(context.Background(), cfg.Name, 10, 0)
	if err != nil {
		t.Fatalf("list by channel failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 message in store, got %d", len(list))
	}
	meta := list[0]
	if meta.ChannelID != cfg.Name {
		t.Errorf("channel id = %q, want %q", meta.ChannelID, cfg.Name)
	}
	if meta.Status != "received" {
		t.Errorf("status = %q, want %q", meta.Status, "received")
	}
	foundMeta = true
	if !foundMeta {
		t.Error("message metadata not found in store")
	}

	// Assert logs contain no payload bytes
	logs := logBuf.String()
	if strings.Contains(logs, "SYNTHETIC_MRN_123456") {
		t.Errorf("logs contained payload bytes (SYNTHETIC_MRN_123456): %s", logs)
	}
	if strings.Contains(logs, "SYNTHETIC_PATIENT") {
		t.Errorf("logs contained payload bytes (SYNTHETIC_PATIENT): %s", logs)
	}
	if strings.Contains(logs, "TEST") && strings.Contains(logs, "SYNTHETIC") {
		// This is a rough check; the key point is no MRN or patient name in logs.
	}
}

func TestLoadConfig(t *testing.T) {
	yamlData := `name: demo-channel
source:
  type: mllp
  host: 127.0.0.1
  port: 2575
destination:
  type: http
  url: http://localhost:8080/webhook
mapping:
  messageType: ADT_A01
  fields:
    - source: PID-3.1
      target: patient_mrn
      transform: copy
`
	tmpDir := t.TempDir()
	path := tmpDir + "/channel.yaml"
	if err := os.WriteFile(path, []byte(yamlData), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if cfg.Name != "demo-channel" {
		t.Errorf("name = %q, want %q", cfg.Name, "demo-channel")
	}
	if cfg.Source.Host != "127.0.0.1" {
		t.Errorf("host = %q, want %q", cfg.Source.Host, "127.0.0.1")
	}
	if cfg.Source.Port != 2575 {
		t.Errorf("port = %d, want %d", cfg.Source.Port, 2575)
	}
	if cfg.Destination.URL != "http://localhost:8080/webhook" {
		t.Errorf("url = %q, want %q", cfg.Destination.URL, "http://localhost:8080/webhook")
	}
	if len(cfg.Mapping.Fields) != 1 {
		t.Fatalf("expected 1 mapping, got %d", len(cfg.Mapping.Fields))
	}
	if cfg.Mapping.Fields[0].Source != "PID-3.1" {
		t.Errorf("mapping source = %q, want %q", cfg.Mapping.Fields[0].Source, "PID-3.1")
	}
}

func TestLoadConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "missing name",
			content: "source:\n  type: mllp\n  host: 127.0.0.1\n  port: 2575\ndestination:\n  type: http\n  url: http://example.com\n",
			wantErr: "channel name is required",
		},
		{
			name:    "unsupported source type",
			content: "name: test\nsource:\n  type: ftp\n  host: 127.0.0.1\n  port: 21\ndestination:\n  type: http\n  url: http://example.com\n",
			wantErr: "unsupported source type",
		},
		{
			name:    "unsupported destination type",
			content: "name: test\nsource:\n  type: mllp\n  host: 127.0.0.1\n  port: 2575\ndestination:\n  type: ftp\n  url: ftp://example.com\n",
			wantErr: "unsupported destination type",
		},
		{
			name:    "missing destination URL",
			content: "name: test\nsource:\n  type: mllp\n  host: 127.0.0.1\n  port: 2575\ndestination:\n  type: http\n",
			wantErr: "destination URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := t.TempDir() + "/channel.yaml"
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("write temp file: %v", err)
			}
			_, err := LoadConfig(path)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}
