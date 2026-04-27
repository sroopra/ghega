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

	result, err := DiffLocal(chPath, store)
	if err != nil {
		t.Fatalf("DiffLocal failed: %v", err)
	}
	if result.Identical {
		t.Error("expected Identical = false when no deployed version exists")
	}
	if result.DeployedHash != "" {
		t.Errorf("DeployedHash = %q, want empty", result.DeployedHash)
	}
	if result.LocalHash == "" {
		t.Error("LocalHash is empty")
	}
	if result.LocalYAML == "" {
		t.Error("LocalYAML is empty")
	}
}

func TestDiffLocal_Identical(t *testing.T) {
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

	result, err := DiffLocal(chPath, store)
	if err != nil {
		t.Fatalf("DiffLocal failed: %v", err)
	}
	if !result.Identical {
		t.Error("expected Identical = true")
	}
	if result.LocalHash != result.DeployedHash {
		t.Errorf("hashes differ: local=%q deployed=%q", result.LocalHash, result.DeployedHash)
	}
	if result.LocalYAML == "" {
		t.Error("LocalYAML is empty")
	}
	if result.DeployedYAML == "" {
		t.Error("DeployedYAML is empty")
	}
}

func TestDiffLocal_Different(t *testing.T) {
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

	result, err := DiffLocal(chPath, store)
	if err != nil {
		t.Fatalf("DiffLocal failed: %v", err)
	}
	if result.Identical {
		t.Error("expected Identical = false")
	}
	if result.LocalHash == result.DeployedHash {
		t.Error("expected different hashes")
	}
	if result.DeployedYAML == "" {
		t.Error("DeployedYAML is empty")
	}
}

func TestDiffLocal_InvalidChannel(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: bad name!
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

	_, err := DiffLocal(chPath, store)
	if err == nil {
		t.Fatal("expected error for invalid channel")
	}
}
