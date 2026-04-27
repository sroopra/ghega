package channel

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestRollback_ToSpecificHash(t *testing.T) {
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

	if _, err := Deploy(chPath, store); err != nil {
		t.Fatalf("second Deploy failed: %v", err)
	}

	if err := Rollback("adt-a01", first.Hash, store); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// RollbackChannel saves one audit and SaveDeploymentAudit saves another.
	audits, err := store.ListDeploymentAudit(context.Background(), "adt-a01")
	if err != nil {
		t.Fatalf("ListDeploymentAudit failed: %v", err)
	}
	var rollbackCount int
	for _, a := range audits {
		if a.Action == "rollback" {
			rollbackCount++
		}
	}
	if rollbackCount != 2 {
		t.Fatalf("rollback audit count = %d, want 2", rollbackCount)
	}
}

func TestRollback_ToPrevious(t *testing.T) {
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

	if _, err := Deploy(chPath, store); err != nil {
		t.Fatalf("second Deploy failed: %v", err)
	}

	if err := Rollback("adt-a01", "", store); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	revs, err := store.ListChannelRevisions(context.Background(), "adt-a01")
	if err != nil {
		t.Fatalf("ListChannelRevisions failed: %v", err)
	}
	if len(revs) != 2 {
		t.Fatalf("len(revs) = %d, want 2", len(revs))
	}

	// Verify rollback targeted the first hash.
	audits, err := store.ListDeploymentAudit(context.Background(), "adt-a01")
	if err != nil {
		t.Fatalf("ListDeploymentAudit failed: %v", err)
	}
	var lastRollbackHash string
	for _, a := range audits {
		if a.Action == "rollback" {
			lastRollbackHash = a.Hash
		}
	}
	if lastRollbackHash != first.Hash {
		t.Errorf("rollback hash = %q, want %q", lastRollbackHash, first.Hash)
	}
}

func TestRollback_NoPreviousRevision(t *testing.T) {
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

	if _, err := Deploy(chPath, store); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	err := Rollback("adt-a01", "", store)
	if err == nil {
		t.Fatal("expected error when no previous revision exists")
	}
}

func TestRollback_InvalidHash(t *testing.T) {
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

	if _, err := Deploy(chPath, store); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	err := Rollback("adt-a01", "invalid-hash", store)
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
}
