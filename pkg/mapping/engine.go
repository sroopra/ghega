package mapping

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
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
	case TransformCEL:
		val, err := msg.getValue(m.Source)
		if err != nil {
			return "", err
		}
		return evaluateCEL(m.Expression, val)
	default:
		return "", fmt.Errorf("unsupported transform %q", m.Transform)
	}
}

// evaluateCEL compiles and evaluates a CEL expression with source bound to the
// provided string value. The result is returned as a string.
func evaluateCEL(expr, source string) (string, error) {
	if expr == "" {
		return "", fmt.Errorf("empty CEL expression")
	}

	env, err := cel.NewEnv(cel.Variable("source", cel.StringType))
	if err != nil {
		return "", fmt.Errorf("cel env: %w", err)
	}

	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return "", fmt.Errorf("cel compile: %w", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return "", fmt.Errorf("cel program: %w", err)
	}

	out, _, err := prg.Eval(map[string]any{"source": source})
	if err != nil {
		return "", fmt.Errorf("cel eval: %w", err)
	}

	return celValueToString(out)
}

// celValueToString converts a CEL evaluation result to a Go string.
func celValueToString(out ref.Val) (string, error) {
	switch v := out.Value().(type) {
	case string:
		return v, nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return "", fmt.Errorf("unsupported CEL result type %T", v)
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
// Missing segments or out-of-range fields return empty strings (standard HL7 semantics).
func (m *hl7Message) getValue(path string) (string, error) {
	segment, fieldNo, compNo, err := parsePath(path)
	if err != nil {
		return "", err
	}

	for _, seg := range m.segments {
		if seg.name != segment {
			continue
		}
		return seg.getField(fieldNo, compNo)
	}
	return "", nil
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

func (m *hl7Message) hasSegment(name string) bool {
	for _, seg := range m.segments {
		if seg.name == name {
			return true
		}
	}
	return false
}

func (s hl7Segment) getField(fieldNo, compNo int) (string, error) {
	var idx int
	if s.name == "MSH" {
		// MSH-1 is the field separator (not stored in fields slice).
		if fieldNo == 1 {
			return "|", nil
		}
		// fields[1] is MSH-2, fields[2] is MSH-3, etc.
		idx = fieldNo - 1
	} else {
		// fields[0] is segment name; fields[1] is PID-1, fields[2] is PID-2, etc.
		idx = fieldNo
	}

	if idx >= len(s.fields) {
		return "", nil
	}
	val := s.fields[idx]

	if compNo > 0 {
		comps := strings.Split(val, "^")
		cidx := compNo - 1
		if cidx >= len(comps) {
			return "", nil
		}
		return comps[cidx], nil
	}
	return val, nil
}
