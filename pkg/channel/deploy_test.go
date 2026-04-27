package channel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestDeploy_NewChannel(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")

	yaml := `name: test-ch
description: test
source:
  type: mllp
  config:
    port: 2575
destination:
  type: http
  config:
    url: http://example.com/api
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
	if result.Name != "test-ch" {
		t.Errorf("Name = %q, want %q", result.Name, "test-ch")
	}
	if result.Revision != 1 {
		t.Errorf("Revision = %d, want 1", result.Revision)
	}
	if result.Hash == "" {
		t.Error("Hash is empty")
	}
	if result.PreviousHash != "" {
		t.Errorf("PreviousHash = %q, want empty", result.PreviousHash)
	}
}

func TestDeploy_Idempotent(t *testing.T) {
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

	result1, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	result2, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("Deploy second time failed: %v", err)
	}

	if result1.Hash != result2.Hash {
		t.Errorf("hashes differ: %q vs %q", result1.Hash, result2.Hash)
	}
	if result2.Revision != result1.Revision {
		t.Errorf("Revision changed: %d vs %d", result2.Revision, result1.Revision)
	}
}

func TestDeploy_SecondRevision(t *testing.T) {
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

	result2, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("Deploy second time failed: %v", err)
	}

	if result2.Hash == result1.Hash {
		t.Error("expected different hash for changed channel")
	}
	if result2.Revision != 2 {
		t.Errorf("Revision = %d, want 2", result2.Revision)
	}
	if result2.PreviousHash != result1.Hash {
		t.Errorf("PreviousHash = %q, want %q", result2.PreviousHash, result1.Hash)
	}
}

func TestDeploy_InvalidYAML(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")

	if err := os.WriteFile(path, []byte("not: valid: yaml: ["), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := Deploy(path, store)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
