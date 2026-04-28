package channel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
	"github.com/sroopra/ghega/pkg/mapping"
	"gopkg.in/yaml.v3"
)

func TestDiffLocal_NotDeployed(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("diff-test")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	result, err := DiffLocal(path, store)
	if err != nil {
		t.Fatalf("DiffLocal failed: %v", err)
	}
	if result.ChannelName != "diff-test" {
		t.Errorf("ChannelName = %q, want %q", result.ChannelName, "diff-test")
	}
	if result.Identical {
		t.Error("expected Identical = false for never-deployed channel")
	}
	if result.DeployedHash != "" {
		t.Errorf("DeployedHash = %q, want empty", result.DeployedHash)
	}
	if result.LocalYAML == "" {
		t.Error("LocalYAML is empty")
	}
}

func TestDiffLocal_Identical(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("diff-test")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	result, err := DiffLocal(path, store)
	if err != nil {
		t.Fatalf("DiffLocal failed: %v", err)
	}
	if !result.Identical {
		t.Error("expected Identical = true when local matches deployed")
	}
	if result.LocalHash != result.DeployedHash {
		t.Errorf("LocalHash %q != DeployedHash %q", result.LocalHash, result.DeployedHash)
	}
}

func TestDiffLocal_Different(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("diff-test")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	// Change the local channel.
	ch.Mappings = append(ch.Mappings, mapping.Mapping{Source: "MSH-9.1", Target: "message_type"})
	path = writeTestChannel(t, dir, "channel.yaml", ch)

	result, err := DiffLocal(path, store)
	if err != nil {
		t.Fatalf("DiffLocal failed: %v", err)
	}
	if result.Identical {
		t.Error("expected Identical = false after local change")
	}
	if result.LocalHash == result.DeployedHash {
		t.Error("expected different hashes")
	}
	if result.DeployedYAML == "" {
		t.Error("DeployedYAML is empty")
	}
}

func TestDiffLocal_InvalidYAML(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")
	if err := os.WriteFile(path, []byte("not: valid: yaml: ["), 0644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	_, err := DiffLocal(path, store)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestDiffLocal_ReturnsYAMLStrings(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("diff-test")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	result, err := DiffLocal(path, store)
	if err != nil {
		t.Fatalf("DiffLocal failed: %v", err)
	}

	var localCh Channel
	if err := yaml.Unmarshal([]byte(result.LocalYAML), &localCh); err != nil {
		t.Errorf("LocalYAML is not valid YAML: %v", err)
	}
	var deployedCh Channel
	if err := yaml.Unmarshal([]byte(result.DeployedYAML), &deployedCh); err != nil {
		t.Errorf("DeployedYAML is not valid YAML: %v", err)
	}
}
