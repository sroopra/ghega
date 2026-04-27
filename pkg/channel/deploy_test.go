package channel

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestDeploy_FirstDeploy(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: adt-a01
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	result, err := Deploy(chPath, store)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if result.Name != "adt-a01" {
		t.Errorf("Name = %q, want %q", result.Name, "adt-a01")
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

	audits, err := store.ListDeploymentAudit(context.Background(), "adt-a01")
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

func TestDeploy_SecondDeploy(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: adt-a01
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	first, err := Deploy(chPath, store)
	if err != nil {
		t.Fatalf("first Deploy failed: %v", err)
	}

	// Modify the channel to get a different hash.
	chYAML2 := `name: adt-a01
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_id
    transform: copy
`
	if err := os.WriteFile(chPath, []byte(chYAML2), 0644); err != nil {
		t.Fatalf("write channel 2: %v", err)
	}

	second, err := Deploy(chPath, store)
	if err != nil {
		t.Fatalf("second Deploy failed: %v", err)
	}
	if second.Revision != 2 {
		t.Errorf("Revision = %d, want 2", second.Revision)
	}
	if second.PreviousHash != first.Hash {
		t.Errorf("PreviousHash = %q, want %q", second.PreviousHash, first.Hash)
	}
	if second.Hash == first.Hash {
		t.Error("expected different hash for second deploy")
	}

	audits, err := store.ListDeploymentAudit(context.Background(), "adt-a01")
	if err != nil {
		t.Fatalf("ListDeploymentAudit failed: %v", err)
	}
	if len(audits) != 2 {
		t.Fatalf("len(audits) = %d, want 2", len(audits))
	}
}

func TestDeploy_Idempotent(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: adt-a01
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	first, err := Deploy(chPath, store)
	if err != nil {
		t.Fatalf("first Deploy failed: %v", err)
	}

	second, err := Deploy(chPath, store)
	if err != nil {
		t.Fatalf("second Deploy failed: %v", err)
	}
	if second.Revision != first.Revision {
		t.Errorf("Revision = %d, want %d (idempotent)", second.Revision, first.Revision)
	}
	if second.Hash != first.Hash {
		t.Error("expected same hash for idempotent deploy")
	}
	if second.PreviousHash != first.PreviousHash {
		t.Errorf("PreviousHash changed: %q vs %q", second.PreviousHash, first.PreviousHash)
	}

	// Only one audit should exist because the second deploy is a no-op.
	audits, err := store.ListDeploymentAudit(context.Background(), "adt-a01")
	if err != nil {
		t.Fatalf("ListDeploymentAudit failed: %v", err)
	}
	if len(audits) != 1 {
		t.Fatalf("len(audits) = %d, want 1 (idempotent)", len(audits))
	}
}

func TestDeploy_InvalidChannel(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: invalid_name!
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	_, err := Deploy(chPath, store)
	if err == nil {
		t.Fatal("expected error for invalid channel")
	}
}
