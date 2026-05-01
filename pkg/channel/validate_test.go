package channel

import (
	"strings"
	"testing"

	"github.com/sroopra/ghega/pkg/mapping"
	"gopkg.in/yaml.v3"
)

func makeValidChannel() Channel {
	return Channel{
		Name:        "adt-a01",
		Description: "ADT A01 feed",
		Source: Source{
			Type: "mllp",
			Config: map[string]any{
				"port": 2575,
			},
		},
		Destination: Destination{
			Type: "http",
			Config: map[string]any{
				"url": "http://example.com/api",
			},
		},
		Mappings: []mapping.Mapping{
			{Source: "PID-3.1", Target: "patient_mrn"},
			{Source: "PID-5.1", Target: "last_name"},
		},
		Tests: []Test{
			{
				Name:     "basic",
				Input:    "testdata/basic.hl7",
				Expected: map[string]string{"patient_mrn": "123"},
			},
		},
		Policies: Policies{
			Network: NetworkPolicy{AllowedHosts: []string{"example.com"}},
			Payload: PayloadPolicy{MaxSizeBytes: 1024 * 1024},
			Time:    TimePolicy{MaxProcessingSeconds: 30},
		},
	}
}

func errorFields(errs []ValidationError) []string {
	fields := make([]string, len(errs))
	for i, e := range errs {
		fields[i] = e.Field
	}
	return fields
}

func hasErrorField(errs []ValidationError, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}

func TestValidateYAML_ValidChannel(t *testing.T) {
	ch := makeValidChannel()
	data, err := yaml.Marshal(ch)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	parsed, errs := ValidateYAML(data)
	if len(errs) > 0 {
		t.Fatalf("expected no validation errors, got: %v", errs)
	}
	if parsed == nil {
		t.Fatal("expected parsed channel, got nil")
	}
	if parsed.Name != ch.Name {
		t.Errorf("Name = %q, want %q", parsed.Name, ch.Name)
	}
}

func TestValidateYAML_InvalidYAML(t *testing.T) {
	data := []byte("not: valid: yaml: [")
	parsed, errs := ValidateYAML(data)
	if parsed != nil {
		t.Error("expected nil channel for invalid YAML")
	}
	if len(errs) == 0 {
		t.Fatal("expected validation errors for invalid YAML")
	}
	if !strings.Contains(errs[0].Message, "invalid YAML") {
		t.Errorf("expected error message to contain 'invalid YAML', got: %q", errs[0].Message)
	}
}

func TestValidateYAML_MissingName(t *testing.T) {
	ch := makeValidChannel()
	ch.Name = ""
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "name") {
		t.Errorf("expected error on field 'name', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_InvalidName(t *testing.T) {
	ch := makeValidChannel()
	ch.Name = "ADT_A01"
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "name") {
		t.Errorf("expected error on field 'name', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_MissingSourceType(t *testing.T) {
	ch := makeValidChannel()
	ch.Source.Type = ""
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "source.type") {
		t.Errorf("expected error on field 'source.type', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_InvalidSourceType(t *testing.T) {
	ch := makeValidChannel()
	ch.Source.Type = "kafka"
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "source.type") {
		t.Errorf("expected error on field 'source.type', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_MissingDestinationType(t *testing.T) {
	ch := makeValidChannel()
	ch.Destination.Type = ""
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "destination.type") {
		t.Errorf("expected error on field 'destination.type', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_InvalidDestinationType(t *testing.T) {
	ch := makeValidChannel()
	ch.Destination.Type = "mllp"
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "destination.type") {
		t.Errorf("expected error on field 'destination.type', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_EmptyMappings(t *testing.T) {
	ch := makeValidChannel()
	ch.Mappings = []mapping.Mapping{}
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "mappings") {
		t.Errorf("expected error on field 'mappings', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_MissingMappingSource(t *testing.T) {
	ch := makeValidChannel()
	ch.Mappings = []mapping.Mapping{{Source: "", Target: "patient_mrn"}}
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "mappings[0].source") {
		t.Errorf("expected error on field 'mappings[0].source', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_MissingMappingTarget(t *testing.T) {
	ch := makeValidChannel()
	ch.Mappings = []mapping.Mapping{{Source: "PID-3.1", Target: ""}}
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "mappings[0].target") {
		t.Errorf("expected error on field 'mappings[0].target', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_InvalidMappingSource(t *testing.T) {
	ch := makeValidChannel()
	ch.Mappings = []mapping.Mapping{{Source: "pid-3.1", Target: "patient_mrn"}}
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "mappings[0].source") {
		t.Errorf("expected error on field 'mappings[0].source', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_InvalidMappingSourceNoComponent(t *testing.T) {
	ch := makeValidChannel()
	ch.Mappings = []mapping.Mapping{{Source: "PID-3", Target: "patient_mrn"}}
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if len(errs) > 0 {
		t.Errorf("expected no errors for PID-3 (field-only is valid), got: %v", errs)
	}
}

func TestValidateYAML_MissingTestName(t *testing.T) {
	ch := makeValidChannel()
	ch.Tests = []Test{{Name: "", Input: "input.hl7"}}
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "tests[0].name") {
		t.Errorf("expected error on field 'tests[0].name', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_MissingTestInput(t *testing.T) {
	ch := makeValidChannel()
	ch.Tests = []Test{{Name: "basic", Input: ""}}
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if !hasErrorField(errs, "tests[0].input") {
		t.Errorf("expected error on field 'tests[0].input', got: %v", errorFields(errs))
	}
}

func TestValidateYAML_NoTestsIsValid(t *testing.T) {
	ch := makeValidChannel()
	ch.Tests = nil
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if len(errs) > 0 {
		t.Errorf("expected no errors when tests are absent, got: %v", errs)
	}
}

func TestValidateYAML_MultipleErrors(t *testing.T) {
	ch := makeValidChannel()
	ch.Name = ""
	ch.Source.Type = ""
	ch.Mappings = nil
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if len(errs) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateYAML_FHIRTypesAreValid(t *testing.T) {
	ch := makeValidChannel()
	ch.Source.Type = "fhir"
	ch.Destination.Type = "fhir"
	data, _ := yaml.Marshal(ch)
	_, errs := ValidateYAML(data)
	if hasErrorField(errs, "source.type") {
		t.Errorf("expected source.type fhir to be valid, got error: %v", errs)
	}
	if hasErrorField(errs, "destination.type") {
		t.Errorf("expected destination.type fhir to be valid, got error: %v", errs)
	}
}
