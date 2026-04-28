package channel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
)

func makeTempChannel(t *testing.T, dir, name string, extra string) string {
	t.Helper()
	chYAML := `name: ` + name + `
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
` + extra
	path := filepath.Join(dir, name+".yaml")
	if err := os.WriteFile(path, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}
	return path
}

func TestDeploy_NewChannel(t *testing.T) {
	dir := t.TempDir()
	store := channelstore.NewInMemoryStore()
	path := makeTempChannel(t, dir, "test-ch", "")

	res, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("deploy: %v", err)
	}
	if res.Name != "test-ch" {
		t.Errorf("name = %q, want test-ch", res.Name)
	}
	if res.Hash == "" {
		t.Error("hash is empty")
	}
	if res.Revision != 1 {
		t.Errorf("revision = %d, want 1", res.Revision)
	}
	if res.PreviousHash != "" {
		t.Errorf("previousHash = %q, want empty", res.PreviousHash)
	}
}

func TestDeploy_Idempotent(t *testing.T) {
	dir := t.TempDir()
	store := channelstore.NewInMemoryStore()
	path := makeTempChannel(t, dir, "test-ch", "")

	res1, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	res2, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	if res2.Revision != res1.Revision {
		t.Errorf("revision changed: %d -> %d", res1.Revision, res2.Revision)
	}
	if res2.Hash != res1.Hash {
		t.Errorf("hash changed: %s -> %s", res1.Hash, res2.Hash)
	}
	if !res2.IsNoOp {
		t.Error("expected IsNoOp to be true")
	}
}

func TestDeploy_PreviousHash(t *testing.T) {
	dir := t.TempDir()
	store := channelstore.NewInMemoryStore()
	path1 := makeTempChannel(t, dir, "test-ch", "")

	res1, err := Deploy(path1, store)
	if err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	path2 := makeTempChannel(t, dir, "test-ch", "  - source: MSH-9.1\n    target: message_type\n")
	res2, err := Deploy(path2, store)
	if err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	if res2.PreviousHash != res1.Hash {
		t.Errorf("previousHash = %q, want %q", res2.PreviousHash, res1.Hash)
	}
	if res2.Revision != 2 {
		t.Errorf("revision = %d, want 2", res2.Revision)
	}
}

func TestDeploy_ValidationError(t *testing.T) {
	dir := t.TempDir()
	store := channelstore.NewInMemoryStore()
	badYAML := `name: bad name
source:
  type: mllp
destination:
  type: http
mappings: []
`
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte(badYAML), 0644)

	_, err := Deploy(path, store)
	if err == nil {
		t.Fatal("expected error for invalid channel")
	}
}
