package channel

import (
	"testing"

	"github.com/sroopra/ghega/pkg/mapping"
)

func newTestChannel(name string) *Channel {
	return &Channel{
		Name:        name,
		Description: "test channel",
		Source: Source{
			Type: "mllp",
			Config: map[string]any{
				"port": 2575,
				"host": "0.0.0.0",
			},
		},
		Destination: Destination{
			Type: "http",
			Config: map[string]any{
				"url":    "http://example.com/api",
				"method": "POST",
			},
		},
		Mappings: []mapping.Mapping{
			{Source: "PID-3.1", Target: "patient_mrn"},
			{Source: "PID-5.1", Target: "last_name"},
			{Source: "MSH-9.1", Target: "message_type"},
		},
		Tests: []Test{
			{
				Name:     "basic",
				Input:    "MSH|^~\\&|SEND|RECV|...",
				Expected: map[string]string{"patient_mrn": "123"},
			},
		},
		Policies: Policies{
			Network: NetworkPolicy{AllowedHosts: []string{"example.com"}},
			Payload: PayloadPolicy{MaxSizeBytes: 1024},
			Time:    TimePolicy{MaxProcessingSeconds: 10},
		},
	}
}

func TestHashChannel_Deterministic(t *testing.T) {
	ch := newTestChannel("adt-a01")

	h1, err := HashChannel(ch)
	if err != nil {
		t.Fatalf("HashChannel failed: %v", err)
	}

	h2, err := HashChannel(ch)
	if err != nil {
		t.Fatalf("HashChannel failed: %v", err)
	}

	if h1 != h2 {
		t.Errorf("same channel produced different hashes: %q vs %q", h1, h2)
	}

	if h1 == "" {
		t.Error("hash is empty")
	}
}

func TestHashChannel_DifferentChannels(t *testing.T) {
	ch1 := newTestChannel("adt-a01")
	ch2 := newTestChannel("adt-a01")
	ch2.Description = "different description"

	h1, err := HashChannel(ch1)
	if err != nil {
		t.Fatalf("HashChannel failed: %v", err)
	}

	h2, err := HashChannel(ch2)
	if err != nil {
		t.Fatalf("HashChannel failed: %v", err)
	}

	if h1 == h2 {
		t.Errorf("different channels produced same hash: %q", h1)
	}
}

func TestHashChannel_MapOrderIndependent(t *testing.T) {
	ch1 := newTestChannel("adt-a01")
	ch1.Source.Config = map[string]any{
		"alpha": 1,
		"beta":  2,
		"gamma": 3,
	}

	ch2 := newTestChannel("adt-a01")
	ch2.Source.Config = map[string]any{
		"gamma": 3,
		"alpha": 1,
		"beta":  2,
	}

	h1, err := HashChannel(ch1)
	if err != nil {
		t.Fatalf("HashChannel failed: %v", err)
	}

	h2, err := HashChannel(ch2)
	if err != nil {
		t.Fatalf("HashChannel failed: %v", err)
	}

	if h1 != h2 {
		t.Errorf("channels with same maps in different order produced different hashes: %q vs %q", h1, h2)
	}
}

func TestHashChannel_Nil(t *testing.T) {
	_, err := HashChannel(nil)
	if err == nil {
		t.Fatal("expected error for nil channel")
	}
}
