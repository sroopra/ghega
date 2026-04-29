package mapping

import (
	"testing"
)

func TestEngineApply(t *testing.T) {
	// Synthetic ADT_A01 message with no PHI.
	raw := []byte(
		"MSH|^~\\&|GhegaApp|GhegaFac|DestApp|DestFac|20240101120000||ADT^A01|MSG001|P|2.5\r" +
			"EVN|A01|20240101120000|\r" +
			"PID|1||MRN12345^^^Hospital^MR||Doe^John^Middle||19800101|M\r",
	)

	tests := []struct {
		name      string
		mappings  []Mapping
		want      map[string]string
		wantErr   bool
		errSubstr string
	}{
		{
			name: "copy PID-3.1 to patient_mrn",
			mappings: []Mapping{
				{Source: "PID-3.1", Target: "patient_mrn", Transform: TransformCopy},
			},
			want: map[string]string{
				"patient_mrn": "MRN12345",
			},
		},
		{
			name: "uppercase PID-5.1 to last_name",
			mappings: []Mapping{
				{Source: "PID-5.1", Target: "last_name", Transform: TransformUppercase},
			},
			want: map[string]string{
				"last_name": "DOE",
			},
		},
		{
			name: "lowercase PID-5.2 to first_name",
			mappings: []Mapping{
				{Source: "PID-5.2", Target: "first_name", Transform: TransformLowercase},
			},
			want: map[string]string{
				"first_name": "john",
			},
		},
		{
			name: "static mapping to source_system",
			mappings: []Mapping{
				{Target: "source_system", Transform: TransformStatic, Value: "ghega-test"},
			},
			want: map[string]string{
				"source_system": "ghega-test",
			},
		},
		{
			name: "multiple mappings",
			mappings: []Mapping{
				{Source: "PID-3.1", Target: "patient_mrn", Transform: TransformCopy},
				{Source: "PID-5.1", Target: "last_name", Transform: TransformUppercase},
				{Target: "source_system", Transform: TransformStatic, Value: "ghega-test"},
			},
			want: map[string]string{
				"patient_mrn":   "MRN12345",
				"last_name":     "DOE",
				"source_system": "ghega-test",
			},
		},
		{
			name: "default transform is copy",
			mappings: []Mapping{
				{Source: "PID-3.1", Target: "patient_mrn"},
			},
			want: map[string]string{
				"patient_mrn": "MRN12345",
			},
		},
		{
			name: "missing segment",
			mappings: []Mapping{
				{Source: "PV1-1", Target: "visit", Transform: TransformCopy},
			},
			want: map[string]string{
				"visit": "",
			},
		},
		{
			name: "unsupported transform",
			mappings: []Mapping{
				{Source: "PID-3.1", Target: "x", Transform: "reverse"},
			},
			wantErr:   true,
			errSubstr: "unsupported transform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng := NewEngine(tt.mappings)
			got, err := eng.Apply(raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
				}
				if tt.errSubstr != "" && !contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %q", tt.errSubstr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d entries, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				if got[k] != wantV {
					t.Errorf("key %q: got %q, want %q", k, got[k], wantV)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseHL7(t *testing.T) {
	raw := []byte("MSH|^~\\&|App|Fac\rPID|1||MRN12345\r")
	msg, err := parseHL7(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msg.segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(msg.segments))
	}
	if msg.segments[0].name != "MSH" {
		t.Errorf("expected first segment MSH, got %s", msg.segments[0].name)
	}
	if msg.segments[1].name != "PID" {
		t.Errorf("expected second segment PID, got %s", msg.segments[1].name)
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		path      string
		seg       string
		field     int
		component int
		wantErr   bool
	}{
		{"PID-3.1", "PID", 3, 1, false},
		{"MSH-9", "MSH", 9, 0, false},
		{"EVN-2", "EVN", 2, 0, false},
		{"PID-3", "PID", 3, 0, false},
		{"BAD", "", 0, 0, true},
		{"PID", "", 0, 0, true},
		{"PID-0", "", 0, 0, true},
		{"PID-3.0", "", 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			seg, field, comp, err := parsePath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for path %q", tt.path)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if seg != tt.seg || field != tt.field || comp != tt.component {
				t.Fatalf("parsePath(%q) = (%q, %d, %d), want (%q, %d, %d)",
					tt.path, seg, field, comp, tt.seg, tt.field, tt.component)
			}
		})
	}
}

func TestMSHFieldOne(t *testing.T) {
	// MSH-1 should resolve to the field separator character.
	raw := []byte("MSH|^~\\&|App|Fac\r")
	msg, err := parseHL7(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := msg.getValue("MSH-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "|" {
		t.Errorf("MSH-1 = %q, want %q", val, "|")
	}
}

func TestEmptyMessage(t *testing.T) {
	eng := NewEngine([]Mapping{{Source: "PID-3.1", Target: "x", Transform: TransformCopy}})
	_, err := eng.Apply([]byte{})
	if err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestEngineApplyCEL(t *testing.T) {
	raw := []byte(
		"MSH|^~\\&|GhegaApp|GhegaFac|DestApp|DestFac|20240101120000||ADT^A01|MSG001|P|2.5\r" +
			"PID|1||MRN12345^^^Hospital^MR||Doe^John^Middle||19800101|M\r",
	)

	tests := []struct {
		name      string
		mappings  []Mapping
		want      map[string]string
		wantErr   bool
		errSubstr string
	}{
		{
			name: "CEL ternary male",
			mappings: []Mapping{
				{Source: "PID-8", Target: "gender", Transform: TransformCEL, Expression: "source == 'M' ? 'Male' : 'Female'"},
			},
			want: map[string]string{
				"gender": "Male",
			},
		},
		{
			name: "CEL ternary female",
			mappings: []Mapping{
				{Source: "PID-8", Target: "gender", Transform: TransformCEL, Expression: "source == 'F' ? 'Female' : 'Other'"},
			},
			want: map[string]string{
				"gender": "Other",
			},
		},
		{
			name: "CEL string concatenation",
			mappings: []Mapping{
				{Source: "PID-5.1", Target: "full_name", Transform: TransformCEL, Expression: "source + ', ' + source"},
			},
			want: map[string]string{
				"full_name": "Doe, Doe",
			},
		},
		{
			name: "CEL arithmetic",
			mappings: []Mapping{
				{Source: "PID-8", Target: "result", Transform: TransformCEL, Expression: "int(source.size()) + 5"},
			},
			want: map[string]string{
				"result": "6",
			},
		},
		{
			name: "CEL with empty source",
			mappings: []Mapping{
				{Source: "PV1-1", Target: "visit", Transform: TransformCEL, Expression: "source == '' ? 'empty' : 'not empty'"},
			},
			want: map[string]string{
				"visit": "empty",
			},
		},
		{
			name: "CEL invalid expression",
			mappings: []Mapping{
				{Source: "PID-8", Target: "gender", Transform: TransformCEL, Expression: "this is not valid"},
			},
			wantErr:   true,
			errSubstr: "cel compile",
		},
		{
			name: "CEL empty expression",
			mappings: []Mapping{
				{Source: "PID-8", Target: "gender", Transform: TransformCEL, Expression: ""},
			},
			wantErr:   true,
			errSubstr: "empty CEL expression",
		},
		{
			name: "CEL type mismatch",
			mappings: []Mapping{
				{Source: "PID-8", Target: "gender", Transform: TransformCEL, Expression: "source + 1"},
			},
			wantErr:   true,
			errSubstr: "cel compile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng := NewEngine(tt.mappings)
			got, err := eng.Apply(raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
				}
				if tt.errSubstr != "" && !contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %q", tt.errSubstr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d entries, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				if got[k] != wantV {
					t.Errorf("key %q: got %q, want %q", k, got[k], wantV)
				}
			}
		})
	}
}

func TestEvaluateCELDirectly(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		source    string
		want      string
		wantErr   bool
		errSubstr string
	}{
		{
			name:   "ternary true",
			expr:   "source == 'M' ? 'Male' : 'Female'",
			source: "M",
			want:   "Male",
		},
		{
			name:   "ternary false",
			expr:   "source == 'M' ? 'Male' : 'Female'",
			source: "F",
			want:   "Female",
		},
		{
			name:   "string concatenation",
			expr:   "'Hello, ' + source + '!'",
			source: "World",
			want:   "Hello, World!",
		},
		{
			name:   "arithmetic addition",
			expr:   "int(source) + 10",
			source: "5",
			want:   "15",
		},
		{
			name:   "arithmetic multiplication",
			expr:   "int(source) * 3",
			source: "4",
			want:   "12",
		},
		{
			name:   "boolean result true",
			expr:   "source == 'active'",
			source: "active",
			want:   "true",
		},
		{
			name:   "boolean result false",
			expr:   "source == 'active'",
			source: "inactive",
			want:   "false",
		},
		{
			name:   "double quoted strings",
			expr:   `source == "M" ? "Male" : "Female"`,
			source: "M",
			want:   "Male",
		},
		{
			name:      "invalid expression",
			expr:      "not a valid cel expression",
			source:    "x",
			wantErr:   true,
			errSubstr: "cel compile",
		},
		{
			name:      "empty expression",
			expr:      "",
			source:    "x",
			wantErr:   true,
			errSubstr: "empty CEL expression",
		},
		{
			name:      "type mismatch",
			expr:      "source + 1",
			source:    "hello",
			wantErr:   true,
			errSubstr: "cel compile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateCEL(tt.expr, tt.source)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
				}
				if tt.errSubstr != "" && !contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %q", tt.errSubstr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
