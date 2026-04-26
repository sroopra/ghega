// Package hl7v2 provides a basic HL7v2 parser and ACK generator.
// This is a minimal implementation sufficient for MLLP integration;
// full parsing will be provided by a dedicated work item.
package hl7v2

import (
	"fmt"
	"strings"
)

// Message represents a parsed HL7v2 message.
type Message struct {
	Raw       []byte
	Segments  []Segment
	Separators Separators
}

// Segment represents a single HL7 segment.
type Segment struct {
	Name   string
	Fields []string
}

// Separators holds the HL7 separator characters from MSH-1 and MSH-2.
type Separators struct {
	Field        string
	Component    string
	Repetition   string
	Escape       string
	Subcomponent string
}

// Parse parses a raw HL7v2 message into a Message.
func Parse(raw []byte) (*Message, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty message")
	}

	// Normalize line endings to \r
	content := strings.ReplaceAll(string(raw), "\n", "\r")
	segments := strings.Split(content, "\r")

	msg := &Message{
		Raw:      raw,
		Segments: make([]Segment, 0, len(segments)),
	}

	for _, segStr := range segments {
		segStr = strings.TrimSpace(segStr)
		if segStr == "" {
			continue
		}
		seg := parseSegment(segStr)
		msg.Segments = append(msg.Segments, seg)

		if seg.Name == "MSH" {
			msg.Separators = parseSeparators(seg)
		}
	}

	if len(msg.Segments) == 0 {
		return nil, fmt.Errorf("no segments found")
	}

	if msg.Segments[0].Name != "MSH" {
		return nil, fmt.Errorf("message must start with MSH segment")
	}

	return msg, nil
}

func parseSegment(segStr string) Segment {
	// MSH is special: the first character after "MSH" is the field separator itself.
	// e.g. MSH|^~\&|APP|...  -> separator is |
	// For non-MSH segments, we just split by |.
	if !strings.HasPrefix(segStr, "MSH") {
		fields := strings.Split(segStr, "|")
		return Segment{Name: fields[0], Fields: fields}
	}

	// For MSH, extract the separator character (the char right after "MSH")
	if len(segStr) < 4 {
		return Segment{Name: "MSH", Fields: []string{"MSH"}}
	}
	sep := string(segStr[3])
	fields := strings.Split(segStr, sep)
	// The first field is "MSH", second is the separator char itself (MSH-1)
	// Re-insert the separator as fields[1] to preserve HL7 field indexing
	expanded := make([]string, 0, len(fields)+1)
	expanded = append(expanded, fields[0]) // "MSH"
	expanded = append(expanded, sep)       // MSH-1: field separator
	expanded = append(expanded, fields[1:]...)
	return Segment{Name: "MSH", Fields: expanded}
}

func parseSeparators(msh Segment) Separators {
	s := Separators{
		Field:        "|",
		Component:    "^",
		Repetition:   "~",
		Escape:       "\\",
		Subcomponent: "&",
	}
	if len(msh.Fields) > 1 && msh.Fields[1] != "" {
		s.Field = msh.Fields[1]
	}
	if len(msh.Fields) > 2 && len(msh.Fields[2]) >= 4 {
		s.Component = string(msh.Fields[2][0])
		s.Repetition = string(msh.Fields[2][1])
		s.Escape = string(msh.Fields[2][2])
		s.Subcomponent = string(msh.Fields[2][3])
	}
	return s
}

// GetField returns the field value at the given segment and position (1-based).
func (m *Message) GetField(segment string, fieldIndex int) string {
	for _, seg := range m.Segments {
		if seg.Name == segment {
			if fieldIndex < len(seg.Fields) {
				return seg.Fields[fieldIndex]
			}
			return ""
		}
	}
	return ""
}

// GenerateACK generates an HL7v2 ACK message for the given incoming message.
func GenerateACK(msg *Message, ackCode string) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	msh := msg.Segments[0]
	if msh.Name != "MSH" {
		return nil, fmt.Errorf("message does not start with MSH")
	}

	sep := msg.Separators
	if sep.Field == "" {
		sep = Separators{Field: "|", Component: "^", Repetition: "~", Escape: "\\", Subcomponent: "&"}
	}

	// Extract values from incoming MSH
	// With proper MSH parsing, fields[1] is the separator, fields[2] is encoding chars.
	sendingApp := safeField(msh, 3)
	sendingFacility := safeField(msh, 4)
	receivingApp := safeField(msh, 5)
	receivingFacility := safeField(msh, 6)
	controlID := safeField(msh, 10)

	// Build ACK MSH: swap sending/receiving apps and facilities
	parts := []string{
		"MSH",
		sep.Component + sep.Repetition + sep.Escape + sep.Subcomponent,
		receivingApp,
		receivingFacility,
		sendingApp,
		sendingFacility,
		"", // timestamp (optional)
		"", // security (optional)
		"ACK",
		controlID,
		"P",
		"2.5",
	}
	ackMSH := strings.Join(parts, sep.Field)

	// Build MSA
	ackMSA := fmt.Sprintf("MSA%s%s%s%s", sep.Field, ackCode, sep.Field, controlID)

	ack := ackMSH + "\r" + ackMSA + "\r"
	return []byte(ack), nil
}

func safeField(seg Segment, idx int) string {
	if idx < len(seg.Fields) {
		return seg.Fields[idx]
	}
	return ""
}
