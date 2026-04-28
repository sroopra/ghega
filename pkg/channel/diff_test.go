package channel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestDiffLocal_NotDeployed(t *testing.T) {
	dir := t.TempDir()
	store := channelstore.NewInMemoryStore()
	path := makeTempChannel(t, dir, "diff-ch", "")

	res, err := DiffLocal(path, store)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if res.Identical {
		t.Error("expected not identical")
	}
	if res.DeployedHash != "" {
		t.Errorf("deployedHash = %q, want empty", res.DeployedHash)
	}
	if res.ChannelName != "diff-ch" {
		t.Errorf("channelName = %q, want diff-ch", res.ChannelName)
	}
}

func TestDiffLocal_Identical(t *testing.T) {
	dir := t.TempDir()
	store := channelstore.NewInMemoryStore()
	path := makeTempChannel(t, dir, "diff-ch", "")

	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	res, err := DiffLocal(path, store)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if !res.Identical {
		t.Error("expected identical")
	}
	if res.LocalHash != res.DeployedHash {
		t.Errorf("hashes differ: local=%s deployed=%s", res.LocalHash, res.DeployedHash)
	}
}

func TestDiffLocal_Different(t *testing.T) {
	dir := t.TempDir()
	store := channelstore.NewInMemoryStore()
	path1 := makeTempChannel(t, dir, "diff-ch", "")

	if _, err := Deploy(path1, store); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	path2 := makeTempChannel(t, dir, "diff-ch", "  - source: MSH-9.1\n    target: message_type\n")
	res, err := DiffLocal(path2, store)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if res.Identical {
		t.Error("expected different")
	}
	if res.LocalHash == res.DeployedHash {
		t.Error("expected different hashes")
	}
}

func TestDiffLocal_ValidationError(t *testing.T) {
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

	_, err := DiffLocal(path, store)
	if err == nil {
		t.Fatal("expected error for invalid channel")
	}
}
