package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sroopra/ghega/pkg/channel"
	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestIntegration_GenerateValidateTestDeployDiffRollback(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()

	// 1. Generate a channel.
	outDir := filepath.Join(dir, "generated")
	if err := runGenerate([]string{"channel", "mllp-to-http", "--name", "integration-test", "--out", outDir}); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	channelPath := filepath.Join(outDir, "channel.yaml")
	if _, err := os.Stat(channelPath); err != nil {
		t.Fatalf("channel.yaml not created: %v", err)
	}

	// 2. Validate it.
	if err := runChannelValidate([]string{channelPath}); err != nil {
		t.Fatalf("validate failed: %v", err)
	}

	// 3. Test it.
	if err := runChannelTest([]string{channelPath}); err != nil {
		t.Fatalf("test failed: %v", err)
	}

	// 4. Deploy it.
	first, err := channel.Deploy(channelPath, store)
	if err != nil {
		t.Fatalf("deploy via package failed: %v", err)
	}

	// 5. Edit the channel (change a mapping target).
	data, err := os.ReadFile(channelPath)
	if err != nil {
		t.Fatalf("read channel: %v", err)
	}
	edited := string(data)
	edited = strings.Replace(edited, "patient_mrn", "patient_id", 1)
	edited = strings.Replace(edited, "patient_mrn: SYNTHETIC_MRN_123456", "patient_id: SYNTHETIC_MRN_123456", 1)
	editPath := filepath.Join(outDir, "edited-channel.yaml")
	if err := os.WriteFile(editPath, []byte(edited), 0644); err != nil {
		t.Fatalf("write edited channel: %v", err)
	}

	// 6. Validate edited channel.
	if err := runChannelValidate([]string{editPath}); err != nil {
		t.Fatalf("validate edited failed: %v", err)
	}

	// 7. Test edited channel.
	start := time.Now()
	if err := runChannelTest([]string{editPath}); err != nil {
		t.Fatalf("test edited failed: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed > 5*time.Second {
		t.Errorf("edit-to-test took %s, want under 5s", elapsed)
	}

	// 8. Diff should detect change (before deploying edited channel).
	diffResult, err := channel.DiffLocal(editPath, store)
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if diffResult.Identical {
		t.Error("expected diff to show changes, but got identical")
	}

	// 9. Deploy edited channel.
	second, err := channel.Deploy(editPath, store)
	if err != nil {
		t.Fatalf("deploy edited failed: %v", err)
	}
	if second.Hash == first.Hash {
		t.Fatal("expected different hash after edit")
	}

	// 10. Rollback.
	if err := channel.Rollback("integration-test", first.Hash, store); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	// 11. Verify rolled-back channel matches original.
	current, err := store.GetChannel(nil, "integration-test")
	if err != nil {
		t.Fatalf("GetChannel after rollback: %v", err)
	}
	if current.Hash != first.Hash {
		t.Errorf("after rollback hash = %q, want %q", current.Hash, first.Hash)
	}
}
