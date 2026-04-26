package hl7v2

import (
	"strings"
	"testing"
)

func makeSyntheticADT() string {
	return "MSH|^~\\&|SENDING_APP|SENDING_FACILITY|RECEIVING_APP|RECEIVING_FACILITY|20240101120000||ADT^A01|MSG001|P|2.5\r" +
		"EVN|A01|20240101120000\r" +
		"PID|1||PAT001^^^MRN||DOE^JOHN||19800101|M\r"
}

func TestParseSyntheticADT(t *testing.T) {
	raw := []byte(makeSyntheticADT())
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(msg.Segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(msg.Segments))
	}
	if msg.Segments[0].Name != "MSH" {
		t.Errorf("expected first segment MSH, got %s", msg.Segments[0].Name)
	}
	if msg.Segments[1].Name != "EVN" {
		t.Errorf("expected second segment EVN, got %s", msg.Segments[1].Name)
	}
	if msg.Segments[2].Name != "PID" {
		t.Errorf("expected third segment PID, got %s", msg.Segments[2].Name)
	}
}

func TestGenerateAA(t *testing.T) {
	raw := []byte(makeSyntheticADT())
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	ack, err := GenerateACK(msg, "AA")
	if err != nil {
		t.Fatalf("generate ack failed: %v", err)
	}
	ackStr := string(ack)
	if !strings.Contains(ackStr, "MSA|AA|MSG001") {
		t.Errorf("ACK missing expected MSA|AA|MSG001, got: %s", ackStr)
	}
	if !strings.HasPrefix(ackStr, "MSH|") {
		t.Errorf("ACK must start with MSH|, got: %s", ackStr)
	}
}

func TestGenerateAR(t *testing.T) {
	raw := []byte(makeSyntheticADT())
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	ack, err := GenerateACK(msg, "AR")
	if err != nil {
		t.Fatalf("generate ack failed: %v", err)
	}
	if !strings.Contains(string(ack), "MSA|AR|MSG001") {
		t.Errorf("ACK missing expected MSA|AR|MSG001")
	}
}

func TestGenerateAE(t *testing.T) {
	raw := []byte(makeSyntheticADT())
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	ack, err := GenerateACK(msg, "AE")
	if err != nil {
		t.Fatalf("generate ack failed: %v", err)
	}
	if !strings.Contains(string(ack), "MSA|AE|MSG001") {
		t.Errorf("ACK missing expected MSA|AE|MSG001")
	}
}

func TestParseEmptyMessage(t *testing.T) {
	_, err := Parse([]byte{})
	if err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestParseNoMSH(t *testing.T) {
	_, err := Parse([]byte("PID|1||PAT001\r"))
	if err == nil {
		t.Fatal("expected error when message does not start with MSH")
	}
}
