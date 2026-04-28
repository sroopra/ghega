package channel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
	"github.com/sroopra/ghega/pkg/mapping"
	"gopkg.in/yaml.v3"
)

func writeTestChannel(t *testing.T, dir, name string, ch *Channel) string {
	t.Helper()
	data, err := yaml.Marshal(ch)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	return p
}

func testChannel(name string) *Channel {
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

func TestDeploy_NewChannel(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("deploy-test")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	result, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if result.Name != "deploy-test" {
		t.Errorf("Name = %q, want %q", result.Name, "deploy-test")
	}
	if result.Hash == "" {
		t.Error("Hash is empty")
	}
	if result.Revision != 1 {
		t.Errorf("Revision = %d, want 1", result.Revision)
	}
	if result.PreviousHash != "" {
		t.Errorf("PreviousHash = %q, want empty", result.PreviousHash)
	}

	audits, err := store.ListDeploymentAudit(nil, "deploy-test")
	if err != nil {
		t.Fatalf("ListDeploymentAudit failed: %v", err)
	}
	if len(audits) != 1 {
		t.Fatalf("len(audits) = %d, want 1", len(audits))
	}
	if audits[0].Action != "deploy" {
		t.Errorf("Action = %q, want %q", audits[0].Action, "deploy")
	}
}

func TestDeploy_SecondRevision(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("deploy-test")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	first, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("first Deploy failed: %v", err)
	}

	// Modify the channel and deploy again.
	ch.Description = "updated description"
	path = writeTestChannel(t, dir, "channel.yaml", ch)

	second, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("second Deploy failed: %v", err)
	}
	if second.Revision != 2 {
		t.Errorf("Revision = %d, want 2", second.Revision)
	}
	if second.PreviousHash != first.Hash {
		t.Errorf("PreviousHash = %q, want %q", second.PreviousHash, first.Hash)
	}
}

func TestDeploy_Idempotent(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("deploy-test")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	first, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("first Deploy failed: %v", err)
	}

	second, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("second Deploy failed: %v", err)
	}
	if second.Revision != first.Revision {
		t.Errorf("Revision = %d, want %d", second.Revision, first.Revision)
	}
	if second.Hash != first.Hash {
		t.Errorf("Hash = %q, want %q", second.Hash, first.Hash)
	}
}

func TestDeploy_InvalidYAML(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")
	if err := os.WriteFile(path, []byte("not: valid: yaml: ["), 0644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	_, err := Deploy(path, store)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestDeploy_MissingName(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	_, err := Deploy(path, store)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}
