package mapping

import (
	"fmt"
	"strings"
)

// Apply runs the engine against the provided raw HL7v2 message and returns the
// transformed output as a map.
func (e *Engine) Apply(raw []byte) (map[string]string, error) {
	msg, err := parseHL7(raw)
	if err != nil {
		return nil, fmt.Errorf("parse hl7: %w", err)
	}

	out := make(map[string]string, len(e.Mappings))
	for _, m := range e.Mappings {
		val, err := e.resolve(msg, m)
		if err != nil {
			return nil, fmt.Errorf("mapping %q -> %q: %w", m.Source, m.Target, err)
		}
		out[m.Target] = val
	}
	return out, nil
}

func (e *Engine) resolve(msg *hl7Message, m Mapping) (string, error) {
	switch m.Transform {
	case TransformStatic, "":
		if m.Transform == TransformStatic {
			return m.Value, nil
		}
		fallthrough
	case TransformCopy, TransformUppercase, TransformLowercase:
		val, err := msg.getValue(m.Source)
		if err != nil {
			return "", err
		}
		switch m.Transform {
		case TransformUppercase:
			return strings.ToUpper(val), nil
		case TransformLowercase:
			return strings.ToLower(val), nil
		default:
			return val, nil
		}
	default:
		return "", fmt.Errorf("unsupported transform %q", m.Transform)
	}
}

// ---------------------------------------------------------------------------
// Minimal HL7v2 parser
// ---------------------------------------------------------------------------

type hl7Message struct {
	segments []hl7Segment
}

type hl7Segment struct {
	name   string
	fields []string // fields[0] is segment name; for MSH fields[1] is encoding chars (MSH-2)
}

func parseHL7(raw []byte) (*hl7Message, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty message")
	}

	// Normalise line endings to \r.
	str := strings.ReplaceAll(string(raw), "\r\n", "\r")
	str = strings.ReplaceAll(str, "\n", "\r")

	parts := strings.Split(str, "\r")
	msg := &hl7Message{segments: make([]hl7Segment, 0, len(parts))}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		seg, err := parseSegment(part)
		if err != nil {
			return nil, err
		}
		msg.segments = append(msg.segments, seg)
	}

	if len(msg.segments) == 0 {
		return nil, fmt.Errorf("no segments found")
	}
	return msg, nil
}

func parseSegment(raw string) (hl7Segment, error) {
	if len(raw) < 3 {
		return hl7Segment{}, fmt.Errorf("segment too short")
	}

	name := raw[:3]
	fields := []string{name}

	if name == "MSH" {
		if len(raw) < 4 {
			return hl7Segment{}, fmt.Errorf("msh segment too short")
		}
		// MSH-1 is the field separator itself (character at index 3).
		sep := string(raw[3])
		// Remainder after MSH|.
		remainder := raw[4:]
		fields = append(fields, strings.Split(remainder, sep)...)
	} else {
		fields = append(fields, strings.Split(raw[4:], "|")...)
	}

	return hl7Segment{name: name, fields: fields}, nil
}

// getValue resolves an HL7 field path such as "PID-3.1".
func (m *hl7Message) getValue(path string) (string, error) {
	segment, fieldNo, compNo, err := parsePath(path)
	if err != nil {
		return "", err
	}

	for _, seg := range m.segments {
		if seg.name != segment {
			continue
		}
		val, err := seg.getField(fieldNo, compNo)
		if err != nil {
			return "", err
		}
		return val, nil
	}
	return "", fmt.Errorf("segment %q not found", segment)
}

// parsePath parses an HL7 path like "PID-3.1" into segment, field, component.
func parsePath(path string) (segment string, field int, component int, err error) {
	// Segment is the first 3 chars.
	if len(path) < 3 {
		return "", 0, 0, fmt.Errorf("invalid path %q", path)
	}
	segment = path[:3]
	rest := path[3:]

	if len(rest) == 0 || rest[0] != '-' {
		return "", 0, 0, fmt.Errorf("invalid path %q: missing field separator", path)
	}
	rest = rest[1:]

	// Split field and optional component.
	var fieldStr, compStr string
	if dot := strings.Index(rest, "."); dot >= 0 {
		fieldStr = rest[:dot]
		compStr = rest[dot+1:]
	} else {
		fieldStr = rest
	}

	field, err = parsePositiveInt(fieldStr)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid field number in path %q", path)
	}
	if compStr != "" {
		component, err = parsePositiveInt(compStr)
		if err != nil {
			return "", 0, 0, fmt.Errorf("invalid component number in path %q", path)
		}
	}
	return segment, field, component, nil
}

func parsePositiveInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty number")
	}
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid digit")
		}
		n = n*10 + int(ch-'0')
	}
	if n == 0 {
		return 0, fmt.Errorf("zero is not a valid 1-based index")
	}
	return n, nil
}

func (s hl7Segment) getField(fieldNo, compNo int) (string, error) {
	var idx int
	if s.name == "MSH" {
		// MSH-1 is the field separator (not stored in fields slice).
		if fieldNo == 1 {
			return "|", nil
		}
		// fields[0] == segment name; fields[1] == MSH-2; therefore MSH-n == fields[n-1].
		idx = fieldNo - 1
	} else {
		idx = fieldNo
	}

	if idx >= len(s.fields) {
		return "", fmt.Errorf("field %d out of range in segment %s", fieldNo, s.name)
	}
	val := s.fields[idx]

	if compNo > 0 {
		comps := strings.Split(val, "^")
		cidx := compNo - 1
		if cidx >= len(comps) {
			return "", fmt.Errorf("component %d out of range in field %d", compNo, fieldNo)
		}
		return comps[cidx], nil
	}
	return val, nil
}
