package channel

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestRollback_ToHash(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")

	yaml := `name: test-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	err = Rollback("test-ch", result.Hash, store)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	audits, err := store.ListDeploymentAudit(context.Background(), "test-ch")
	if err != nil {
		t.Fatalf("ListDeploymentAudit failed: %v", err)
	}
	if len(audits) < 2 {
		t.Fatalf("expected at least 2 audits, got %d", len(audits))
	}
}

func TestRollback_ToPrevious(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")

	yaml1 := `name: test-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	if err := os.WriteFile(path, []byte(yaml1), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result1, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	yaml2 := `name: test-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
  - source: PID-5.1
    target: last_name
`
	if err := os.WriteFile(path, []byte(yaml2), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("Deploy second time failed: %v", err)
	}

	// Rollback without specifying hash should go to previous (result1)
	err = Rollback("test-ch", "", store)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	audits, err := store.ListDeploymentAudit(context.Background(), "test-ch")
	if err != nil {
		t.Fatalf("ListDeploymentAudit failed: %v", err)
	}
	if len(audits) < 3 {
		t.Fatalf("expected at least 3 audits, got %d", len(audits))
	}
	// Most recent audit should be rollback with result1 hash
	if audits[0].Hash != result1.Hash {
		t.Errorf("latest audit Hash = %q, want %q", audits[0].Hash, result1.Hash)
	}
}

func TestRollback_NoPreviousRevision(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")

	yaml := `name: test-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	err := Rollback("test-ch", "", store)
	if err == nil {
		t.Fatal("expected error when no previous revision exists")
	}
}

func TestRollback_InvalidHash(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")

	yaml := `name: test-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	err := Rollback("test-ch", "invalid-hash", store)
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
}
