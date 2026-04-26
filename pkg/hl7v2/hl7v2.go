// Package hl7v2 provides a basic HL7v2 message parser and ACK generator.
package hl7v2

import (
	"fmt"
	"strings"
	"time"
)

// Message represents a parsed HL7v2 message.
type Message struct {
	Segments        []Segment
	FieldSeparator  byte
	ComponentSep    byte
	RepetitionSep   byte
	EscapeChar      byte
	SubcomponentSep byte
}

// Segment represents a single HL7v2 segment.
type Segment struct {
	Type   string
	Fields []string
}

// Parse parses a raw HL7v2 message into a Message.
func Parse(raw []byte) (*Message, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty message")
	}

	// Normalize line endings: treat both \r\n and bare \n as segment
	// separators, mapping them to the standard \r.
	str := string(raw)
	str = strings.ReplaceAll(str, "\r\n", "\r")
	str = strings.ReplaceAll(str, "\n", "\r")

	parts := strings.Split(str, "\r")

	msg := &Message{
		FieldSeparator:  '|',
		ComponentSep:    '^',
		RepetitionSep:   '~',
		EscapeChar:      '\\',
		SubcomponentSep: '&',
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// The first non-empty segment must be MSH; the 4th character is the
		// field separator and must be discovered before splitting fields.
		if len(msg.Segments) == 0 && len(part) >= 4 && strings.HasPrefix(part, "MSH") {
			msg.FieldSeparator = part[3]
		}

		seg, err := parseSegment(part, msg)
		if err != nil {
			return nil, err
		}
		msg.Segments = append(msg.Segments, seg)
	}

	if len(msg.Segments) == 0 {
		return nil, fmt.Errorf("no segments found")
	}

	if msg.Segments[0].Type != "MSH" {
		return nil, fmt.Errorf("first segment must be MSH, got %s", msg.Segments[0].Type)
	}

	return msg, nil
}

func parseSegment(s string, msg *Message) (Segment, error) {
	if len(s) < 3 {
		return Segment{}, fmt.Errorf("segment too short: %q", s)
	}

	parts := strings.Split(s, string(msg.FieldSeparator))
	if len(parts) == 0 {
		return Segment{}, fmt.Errorf("empty segment")
	}

	seg := Segment{
		Type:   parts[0],
		Fields: parts[1:],
	}

	if seg.Type == "MSH" {
		// The field separator itself is MSH-1 and was consumed by Split.
		seg.Fields = append([]string{string(msg.FieldSeparator)}, seg.Fields...)

		// Parse encoding characters from MSH-2 to update message separators.
		if len(seg.Fields) > 1 && len(seg.Fields[1]) >= 4 {
			enc := seg.Fields[1]
			msg.ComponentSep = enc[0]
			msg.RepetitionSep = enc[1]
			msg.EscapeChar = enc[2]
			msg.SubcomponentSep = enc[3]
		}
	}

	return seg, nil
}

// Field returns the nth field of the segment (1-indexed).
func (s *Segment) Field(n int) string {
	if n < 1 || n > len(s.Fields) {
		return ""
	}
	return s.Fields[n-1]
}

// Segment returns the first segment with the given type.
func (m *Message) Segment(name string) *Segment {
	for i := range m.Segments {
		if m.Segments[i].Type == name {
			return &m.Segments[i]
		}
	}
	return nil
}

// GenerateACK generates an HL7v2 ACK response for the given message.
func GenerateACK(msg *Message, ackCode string) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	msh := msg.Segment("MSH")
	if msh == nil {
		return nil, fmt.Errorf("message missing MSH segment")
	}

	switch ackCode {
	case "AA", "AR", "AE":
		// valid
	default:
		return nil, fmt.Errorf("invalid ACK code: %s", ackCode)
	}

	// Copy MSH-3, MSH-4, MSH-5, MSH-6 from original message.
	msh3 := msh.Field(3)
	msh4 := msh.Field(4)
	msh5 := msh.Field(5)
	msh6 := msh.Field(6)
	msh10 := msh.Field(10)
	msh11 := msh.Field(11)
	msh12 := msh.Field(12)

	now := time.Now().Format("20060102150405")

	ackMSH := Segment{
		Type: "MSH",
		Fields: []string{
			string(msg.FieldSeparator),
			string([]byte{msg.ComponentSep, msg.RepetitionSep, msg.EscapeChar, msg.SubcomponentSep}),
			msh3,
			msh4,
			msh5,
			msh6,
			now,
			"",
			"ACK",
			generateControlID(),
			msh11,
			msh12,
		},
	}

	ackMSA := Segment{
		Type: "MSA",
		Fields: []string{
			ackCode,
			msh10,
		},
	}

	ack := &Message{
		Segments:        []Segment{ackMSH, ackMSA},
		FieldSeparator:  msg.FieldSeparator,
		ComponentSep:    msg.ComponentSep,
		RepetitionSep:   msg.RepetitionSep,
		EscapeChar:      msg.EscapeChar,
		SubcomponentSep: msg.SubcomponentSep,
	}

	return ack.Serialize(), nil
}

func generateControlID() string {
	return fmt.Sprintf("ACK-%d", time.Now().UnixNano())
}

// Serialize converts the message back to HL7v2 wire format.
func (m *Message) Serialize() []byte {
	var segments []string
	for _, seg := range m.Segments {
		segments = append(segments, seg.serialize(m.FieldSeparator))
	}
	return []byte(strings.Join(segments, "\r") + "\r")
}

func (s *Segment) serialize(sep byte) string {
	if s.Type == "MSH" {
		// MSH-1 is the separator itself and is stored in Fields[0].
		parts := []string{s.Type}
		parts = append(parts, s.Fields[1:]...)
		return strings.Join(parts, s.Fields[0])
	}
	return s.Type + string(sep) + strings.Join(s.Fields, string(sep))
}
