package mllp

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/sroopra/ghega/pkg/hl7v2"
)

func makeSyntheticADT() []byte {
	return []byte("MSH|^~\\&|SENDING_APP|SENDING_FACILITY|RECEIVING_APP|RECEIVING_FACILITY|20240101120000||ADT^A01|MSG001|P|2.5\r" +
		"EVN|A01|20240101120000\r" +
		"PID|1||PAT001^^^MRN||DOE^JOHN||19800101|M\r")
}

func TestEncodeFrame(t *testing.T) {
	payload := []byte("TEST")
	frame := EncodeFrame(payload)
	if len(frame) != len(payload)+3 {
		t.Fatalf("expected frame length %d, got %d", len(payload)+3, len(frame))
	}
	if frame[0] != StartBlock {
		t.Errorf("expected start block 0x%02X, got 0x%02X", StartBlock, frame[0])
	}
	if frame[len(frame)-2] != EndBlock {
		t.Errorf("expected end block 0x%02X, got 0x%02X", EndBlock, frame[len(frame)-2])
	}
	if frame[len(frame)-1] != CarriageReturn {
		t.Errorf("expected CR 0x%02X, got 0x%02X", CarriageReturn, frame[len(frame)-1])
	}
	if !bytes.Equal(frame[1:len(frame)-2], payload) {
		t.Errorf("payload mismatch")
	}
}

func TestDecodeFrame(t *testing.T) {
	payload := []byte("TEST")
	frame := EncodeFrame(payload)
	decoded, n, err := DecodeFrame(frame)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if n != len(frame) {
		t.Errorf("expected consumed %d, got %d", len(frame), n)
	}
	if !bytes.Equal(decoded, payload) {
		t.Errorf("decoded payload mismatch: got %s, want %s", decoded, payload)
	}
}

func TestDecodeFrameShortBuffer(t *testing.T) {
	_, _, err := DecodeFrame([]byte{StartBlock})
	if err != io.ErrShortBuffer {
		t.Errorf("expected ErrShortBuffer, got %v", err)
	}
}

func TestDecodeFrameMissingStartBlock(t *testing.T) {
	_, _, err := DecodeFrame([]byte("TEST"))
	if err == nil {
		t.Fatal("expected error for missing start block")
	}
}

func TestDecodeFrameIncomplete(t *testing.T) {
	data := []byte{StartBlock, 'T', 'E', 'S', 'T'}
	_, _, err := DecodeFrame(data)
	if err != io.ErrShortBuffer {
		t.Errorf("expected ErrShortBuffer, got %v", err)
	}
}

func TestListenerStartStop(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 0}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ln := NewListener(cfg, nil, logger)

	if err := ln.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	addr := ln.Addr()
	if addr == nil {
		t.Fatal("expected non-nil addr after start")
	}

	if err := ln.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}

func TestListenerReceiveAndACK(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 0}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ln := NewListener(cfg, nil, logger)

	if err := ln.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	defer ln.Stop()

	addr := ln.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	payload := makeSyntheticADT()
	frame := EncodeFrame(payload)
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Read ACK response
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read ack failed: %v", err)
	}

	ackPayload, consumed, err := DecodeFrame(buf[:n])
	if err != nil {
		t.Fatalf("decode ack failed: %v", err)
	}
	if consumed != n {
		t.Errorf("expected to consume entire response, consumed %d of %d", consumed, n)
	}

	ackStr := string(ackPayload)
	if !bytes.Contains(ackPayload, []byte("MSA|AA|MSG001")) {
		t.Errorf("ACK missing expected MSA|AA|MSG001, got: %s", ackStr)
	}
}

func TestListenerCustomHandler(t *testing.T) {
	customCalled := false
	customHandler := func(msg *hl7v2.Message) ([]byte, error) {
		customCalled = true
		return hl7v2.GenerateACK(msg, "AR")
	}

	cfg := Config{Host: "127.0.0.1", Port: 0}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ln := NewListener(cfg, customHandler, logger)

	if err := ln.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	defer ln.Stop()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	payload := makeSyntheticADT()
	frame := EncodeFrame(payload)
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read ack failed: %v", err)
	}

	ackPayload, _, err := DecodeFrame(buf[:n])
	if err != nil {
		t.Fatalf("decode ack failed: %v", err)
	}

	if !customCalled {
		t.Error("expected custom handler to be called")
	}
	if !bytes.Contains(ackPayload, []byte("MSA|AR|MSG001")) {
		t.Errorf("expected AR ack, got: %s", string(ackPayload))
	}
}

func TestListenerMultipleMessages(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 0}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ln := NewListener(cfg, nil, logger)

	if err := ln.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	defer ln.Stop()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	for i := 0; i < 3; i++ {
		payload := []byte("MSH|^~\\&|APP|FAC|RCV|RCVFAC|20240101120000||ADT^A01|MSG00" + string(rune('1'+i)) + "|P|2.5\r")
		frame := EncodeFrame(payload)
		if _, err := conn.Write(frame); err != nil {
			t.Fatalf("write message %d failed: %v", i, err)
		}

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("read ack %d failed: %v", i, err)
		}

		ackPayload, _, err := DecodeFrame(buf[:n])
		if err != nil {
			t.Fatalf("decode ack %d failed: %v", i, err)
		}

		expected := "MSA|AA|MSG00" + string(rune('1'+i))
		if !bytes.Contains(ackPayload, []byte(expected)) {
			t.Errorf("message %d: expected %s in ack, got: %s", i, expected, string(ackPayload))
		}
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("GHEGA_MLLP_HOST", "192.168.1.1")
	t.Setenv("GHEGA_MLLP_PORT", "5555")
	cfg := ConfigFromEnv()
	if cfg.Host != "192.168.1.1" {
		t.Errorf("expected host 192.168.1.1, got %s", cfg.Host)
	}
	if cfg.Port != 5555 {
		t.Errorf("expected port 5555, got %d", cfg.Port)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected default host 0.0.0.0, got %s", cfg.Host)
	}
	if cfg.Port != 2575 {
		t.Errorf("expected default port 2575, got %d", cfg.Port)
	}
}

func TestListenerParseErrorReturnsAR(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 0}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ln := NewListener(cfg, nil, logger)

	if err := ln.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	defer ln.Stop()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Send an invalid HL7 message (no MSH)
	payload := []byte("PID|1||PAT001\r")
	frame := EncodeFrame(payload)
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read ack failed: %v", err)
	}

	ackPayload, _, err := DecodeFrame(buf[:n])
	if err != nil {
		t.Fatalf("decode ack failed: %v", err)
	}

	if !bytes.Contains(ackPayload, []byte("MSA|AR|UNKNOWN")) {
		t.Errorf("expected AR ack for parse error, got: %s", string(ackPayload))
	}
}
