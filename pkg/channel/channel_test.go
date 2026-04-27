package channel

import (
	"testing"

	"github.com/sroopra/ghega/pkg/mapping"
	"gopkg.in/yaml.v3"
)

func TestChannel_YAMLRoundTrip(t *testing.T) {
	ch := Channel{
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

	data, err := yaml.Marshal(ch)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed Channel
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.Name != ch.Name {
		t.Errorf("Name = %q, want %q", parsed.Name, ch.Name)
	}
	if parsed.Description != ch.Description {
		t.Errorf("Description = %q, want %q", parsed.Description, ch.Description)
	}
	if parsed.Source.Type != ch.Source.Type {
		t.Errorf("Source.Type = %q, want %q", parsed.Source.Type, ch.Source.Type)
	}
	if len(parsed.Mappings) != len(ch.Mappings) {
		t.Errorf("len(Mappings) = %d, want %d", len(parsed.Mappings), len(ch.Mappings))
	}
	if len(parsed.Tests) != len(ch.Tests) {
		t.Errorf("len(Tests) = %d, want %d", len(parsed.Tests), len(ch.Tests))
	}
	if parsed.Policies.Payload.MaxSizeBytes != ch.Policies.Payload.MaxSizeBytes {
		t.Errorf("MaxSizeBytes = %d, want %d", parsed.Policies.Payload.MaxSizeBytes, ch.Policies.Payload.MaxSizeBytes)
	}
}
