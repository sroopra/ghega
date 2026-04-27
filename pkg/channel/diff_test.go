package channel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestDiffLocal_NoDeployedVersion(t *testing.T) {
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

	result, err := DiffLocal(path, store)
	if err != nil {
		t.Fatalf("DiffLocal failed: %v", err)
	}
	if result.Identical {
		t.Error("expected Identical = false")
	}
	if result.DeployedHash != "" {
		t.Errorf("DeployedHash = %q, want empty", result.DeployedHash)
	}
	if result.LocalYAML != yaml {
		t.Error("LocalYAML mismatch")
	}
}

func TestDiffLocal_Identical(t *testing.T) {
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

	result, err := DiffLocal(path, store)
	if err != nil {
		t.Fatalf("DiffLocal failed: %v", err)
	}
	if !result.Identical {
		t.Error("expected Identical = true")
	}
	if result.LocalHash != result.DeployedHash {
		t.Errorf("hashes differ: %q vs %q", result.LocalHash, result.DeployedHash)
	}
}

func TestDiffLocal_Different(t *testing.T) {
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

	if _, err := Deploy(path, store); err != nil {
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

	result, err := DiffLocal(path, store)
	if err != nil {
		t.Fatalf("DiffLocal failed: %v", err)
	}
	if result.Identical {
		t.Error("expected Identical = false")
	}
	if result.LocalHash == result.DeployedHash {
		t.Error("expected different hashes")
	}
	if result.LocalYAML != yaml2 {
		t.Error("LocalYAML mismatch")
	}
}
