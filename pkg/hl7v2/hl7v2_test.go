package hl7v2

import (
	"strings"
	"testing"
)

// syntheticADT_A01 returns a completely synthetic ADT^A01 message with no
// real patient health information.
func syntheticADT_A01() []byte {
	return []byte(
		"MSH|^~\\&|GHEGA_APP|GHEGA_FACILITY|RECV_APP|RECV_FACILITY|20240101120000||ADT^A01|MSG001|P|2.5\r" +
			"EVN|A01|20240101120000\r" +
			"PID|1||SYNTH001||DOE^JOHN||19800101|M|||123 SYNTH ST^^SYNTHCITY^ST^12345\r" +
			"PV1|1|I|WARD1^^^SYNTH_HOSPITAL\r",
	)
}

func TestParseADT_A01(t *testing.T) {
	raw := syntheticADT_A01()
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(msg.Segments) != 4 {
		t.Fatalf("expected 4 segments, got %d", len(msg.Segments))
	}

	msh := msg.Segment("MSH")
	if msh == nil {
		t.Fatal("missing MSH segment")
	}
	if got, want := msh.Field(1), "|"; got != want {
		t.Errorf("MSH-1 = %q, want %q", got, want)
	}
	if got, want := msh.Field(2), "^~\\&"; got != want {
		t.Errorf("MSH-2 = %q, want %q", got, want)
	}
	if got, want := msh.Field(3), "GHEGA_APP"; got != want {
		t.Errorf("MSH-3 = %q, want %q", got, want)
	}
	if got, want := msh.Field(4), "GHEGA_FACILITY"; got != want {
		t.Errorf("MSH-4 = %q, want %q", got, want)
	}
	if got, want := msh.Field(5), "RECV_APP"; got != want {
		t.Errorf("MSH-5 = %q, want %q", got, want)
	}
	if got, want := msh.Field(6), "RECV_FACILITY"; got != want {
		t.Errorf("MSH-6 = %q, want %q", got, want)
	}
	if got, want := msh.Field(9), "ADT^A01"; got != want {
		t.Errorf("MSH-9 = %q, want %q", got, want)
	}
	if got, want := msh.Field(10), "MSG001"; got != want {
		t.Errorf("MSH-10 = %q, want %q", got, want)
	}

	pid := msg.Segment("PID")
	if pid == nil {
		t.Fatal("missing PID segment")
	}
	if got, want := pid.Field(1), "1"; got != want {
		t.Errorf("PID-1 = %q, want %q", got, want)
	}
	if got, want := pid.Field(3), "SYNTH001"; got != want {
		t.Errorf("PID-3 = %q, want %q", got, want)
	}
	if got, want := pid.Field(5), "DOE^JOHN"; got != want {
		t.Errorf("PID-5 = %q, want %q", got, want)
	}

	// Verify separators were parsed from MSH-2.
	if msg.FieldSeparator != '|' {
		t.Errorf("field separator = %q, want %q", msg.FieldSeparator, '|')
	}
	if msg.ComponentSep != '^' {
		t.Errorf("component separator = %q, want %q", msg.ComponentSep, '^')
	}
	if msg.RepetitionSep != '~' {
		t.Errorf("repetition separator = %q, want %q", msg.RepetitionSep, '~')
	}
	if msg.EscapeChar != '\\' {
		t.Errorf("escape char = %q, want %q", msg.EscapeChar, '\\')
	}
	if msg.SubcomponentSep != '&' {
		t.Errorf("subcomponent separator = %q, want %q", msg.SubcomponentSep, '&')
	}
}

func TestParseCRLF(t *testing.T) {
	// Some systems send \r\n instead of bare \r.
	raw := []byte("MSH|^~\\&|GHEGA_APP|GHEGA_FACILITY|20240101120000||ADT^A01|MSG001|P|2.5\r\nPID|1||SYNTH001\r\n")
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(msg.Segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(msg.Segments))
	}
}

func TestParseCustomSeparators(t *testing.T) {
	// Message with non-standard separators: # field, @ component, ! repetition,
	// $ escape, % subcomponent.
	raw := []byte("MSH#@!$%#GHEGA_APP#GHEGA_FACILITY#20240101120000###ADT@A01#MSG001#P#2.5\r")
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if msg.FieldSeparator != '#' {
		t.Errorf("field separator = %q, want %q", msg.FieldSeparator, '#')
	}
	if msg.ComponentSep != '@' {
		t.Errorf("component separator = %q, want %q", msg.ComponentSep, '@')
	}
	if msg.RepetitionSep != '!' {
		t.Errorf("repetition separator = %q, want %q", msg.RepetitionSep, '!')
	}
	if msg.EscapeChar != '$' {
		t.Errorf("escape char = %q, want %q", msg.EscapeChar, '$')
	}
	if msg.SubcomponentSep != '%' {
		t.Errorf("subcomponent separator = %q, want %q", msg.SubcomponentSep, '%')
	}
}

func TestGenerateACK_AA(t *testing.T) {
	raw := syntheticADT_A01()
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	ack, err := GenerateACK(msg, "AA")
	if err != nil {
		t.Fatalf("generate ack failed: %v", err)
	}

	ackMsg, err := Parse(ack)
	if err != nil {
		t.Fatalf("parse ack failed: %v", err)
	}

	msh := ackMsg.Segment("MSH")
	if msh == nil {
		t.Fatal("missing MSH in ACK")
	}
	if got, want := msh.Field(3), "GHEGA_APP"; got != want {
		t.Errorf("ACK MSH-3 = %q, want %q", got, want)
	}
	if got, want := msh.Field(4), "GHEGA_FACILITY"; got != want {
		t.Errorf("ACK MSH-4 = %q, want %q", got, want)
	}
	if got, want := msh.Field(5), "RECV_APP"; got != want {
		t.Errorf("ACK MSH-5 = %q, want %q", got, want)
	}
	if got, want := msh.Field(6), "RECV_FACILITY"; got != want {
		t.Errorf("ACK MSH-6 = %q, want %q", got, want)
	}
	if got, want := msh.Field(9), "ACK"; got != want {
		t.Errorf("ACK MSH-9 = %q, want %q", got, want)
	}

	msa := ackMsg.Segment("MSA")
	if msa == nil {
		t.Fatal("missing MSA in ACK")
	}
	if got, want := msa.Field(1), "AA"; got != want {
		t.Errorf("MSA-1 = %q, want %q", got, want)
	}
	if got, want := msa.Field(2), "MSG001"; got != want {
		t.Errorf("MSA-2 = %q, want %q", got, want)
	}
}

func TestGenerateACK_AR(t *testing.T) {
	raw := syntheticADT_A01()
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	ack, err := GenerateACK(msg, "AR")
	if err != nil {
		t.Fatalf("generate ack failed: %v", err)
	}

	ackMsg, err := Parse(ack)
	if err != nil {
		t.Fatalf("parse ack failed: %v", err)
	}

	msa := ackMsg.Segment("MSA")
	if msa == nil {
		t.Fatal("missing MSA in ACK")
	}
	if got, want := msa.Field(1), "AR"; got != want {
		t.Errorf("MSA-1 = %q, want %q", got, want)
	}
	if got, want := msa.Field(2), "MSG001"; got != want {
		t.Errorf("MSA-2 = %q, want %q", got, want)
	}
}

func TestGenerateACK_AE(t *testing.T) {
	raw := syntheticADT_A01()
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	ack, err := GenerateACK(msg, "AE")
	if err != nil {
		t.Fatalf("generate ack failed: %v", err)
	}

	ackMsg, err := Parse(ack)
	if err != nil {
		t.Fatalf("parse ack failed: %v", err)
	}

	msa := ackMsg.Segment("MSA")
	if msa == nil {
		t.Fatal("missing MSA in ACK")
	}
	if got, want := msa.Field(1), "AE"; got != want {
		t.Errorf("MSA-1 = %q, want %q", got, want)
	}
	if got, want := msa.Field(2), "MSG001"; got != want {
		t.Errorf("MSA-2 = %q, want %q", got, want)
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse([]byte(""))
	if err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestParseNoMSH(t *testing.T) {
	_, err := Parse([]byte("PID|1||SYNTH001\r"))
	if err == nil {
		t.Fatal("expected error for message without MSH")
	}
}

func TestParseMalformedSegment(t *testing.T) {
	_, err := Parse([]byte("MSH|^~\\&|APP|FAC|20240101120000||ADT^A01|MSG001|P|2.5\rX\r"))
	if err == nil {
		t.Fatal("expected error for malformed segment")
	}
}

func TestGenerateACKInvalidCode(t *testing.T) {
	raw := syntheticADT_A01()
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	_, err = GenerateACK(msg, "XX")
	if err == nil {
		t.Fatal("expected error for invalid ACK code")
	}
}

func TestGenerateACKNilMessage(t *testing.T) {
	_, err := GenerateACK(nil, "AA")
	if err == nil {
		t.Fatal("expected error for nil message")
	}
}

func TestGenerateACKNoMSH(t *testing.T) {
	msg := &Message{
		Segments: []Segment{{Type: "PID", Fields: []string{"1"}}},
	}
	_, err := GenerateACK(msg, "AA")
	if err == nil {
		t.Fatal("expected error for message without MSH")
	}
}

func TestRoundTrip(t *testing.T) {
	raw := syntheticADT_A01()
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	serialized := msg.Serialize()
	msg2, err := Parse(serialized)
	if err != nil {
		t.Fatalf("re-parse failed: %v", err)
	}

	if len(msg.Segments) != len(msg2.Segments) {
		t.Fatalf("segment count mismatch: %d vs %d", len(msg.Segments), len(msg2.Segments))
	}

	for i := range msg.Segments {
		if msg.Segments[i].Type != msg2.Segments[i].Type {
			t.Errorf("segment %d type mismatch: %q vs %q", i, msg.Segments[i].Type, msg2.Segments[i].Type)
		}
		if len(msg.Segments[i].Fields) != len(msg2.Segments[i].Fields) {
			t.Errorf("segment %d field count mismatch: %d vs %d", i, len(msg.Segments[i].Fields), len(msg2.Segments[i].Fields))
			continue
		}
		for j := range msg.Segments[i].Fields {
			if msg.Segments[i].Fields[j] != msg2.Segments[i].Fields[j] {
				t.Errorf("segment %d field %d mismatch: %q vs %q", i, j, msg.Segments[i].Fields[j], msg2.Segments[i].Fields[j])
			}
		}
	}
}

func TestSerializeTrailingEmptyFields(t *testing.T) {
	msg := &Message{
		Segments: []Segment{
			{Type: "MSH", Fields: []string{"|", "^~\\&", "APP", "", "DEST", "", "20240101120000", "", "ACK", "CTRL", "P", "2.5"}},
			{Type: "MSA", Fields: []string{"AA", "CTRL"}},
		},
		FieldSeparator: '|',
	}
	out := string(msg.Serialize())
	if !strings.HasPrefix(out, "MSH|^~\\&|APP||DEST||20240101120000||ACK|CTRL|P|2.5\r") {
		t.Errorf("unexpected serialized MSH: %q", out)
	}
	if !strings.Contains(out, "MSA|AA|CTRL\r") {
		t.Errorf("unexpected serialized MSA: %q", out)
	}
}
