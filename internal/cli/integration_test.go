package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sroopra/ghega/internal/cli/generate"
	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestEndToEnd_EditToTestUnder5Seconds(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "generated-channel")

	// a. Generate a channel to a temp directory.
	if err := generate.RunChannelGenerate([]string{"mllp-to-http", "--name", "e2e-demo", "--message-type", "ADT_A01", "--out", outDir}); err != nil {
		t.Fatalf("generate channel: %v", err)
	}

	channelPath := filepath.Join(outDir, "channel.yaml")
	originalData, err := os.ReadFile(channelPath)
	if err != nil {
		t.Fatalf("read original channel: %v", err)
	}

	// b. Validate it with ghega channel validate.
	if err := runChannelValidate([]string{channelPath}); err != nil {
		t.Fatalf("validate original: %v", err)
	}

	// c. Test it with ghega channel test.
	if err := runChannelTest([]string{channelPath}); err != nil {
		t.Fatalf("test original: %v", err)
	}

	// d. Deploy it with ghega channel deploy.
	if err := runChannelDeployWithStore([]string{channelPath}, store); err != nil {
		t.Fatalf("deploy original: %v", err)
	}

	// e. Verify deployment via store.
	ctx := context.Background()
	originalRecord, err := store.GetChannel(ctx, "e2e-demo")
	if err != nil {
		t.Fatalf("get deployed channel: %v", err)
	}
	if originalRecord == nil {
		t.Fatal("expected deployed channel record, got nil")
	}
	if string(originalRecord.YAML) != string(originalData) {
		t.Error("deployed YAML does not match original")
	}

	// h. Test the edited channel.
	// Measure the time from step (f) edit to step (h) test completion.
	editToTestStart := time.Now()

	// f. Edit the channel YAML (e.g., change a mapping target).
	editedData := strings.ReplaceAll(string(originalData), "patient_mrn", "patient_id")
	if err := os.WriteFile(channelPath, []byte(editedData), 0644); err != nil {
		t.Fatalf("write edited channel: %v", err)
	}

	// g. Validate the edited channel.
	if err := runChannelValidate([]string{channelPath}); err != nil {
		t.Fatalf("validate edited: %v", err)
	}

	if err := runChannelTest([]string{channelPath}); err != nil {
		t.Fatalf("test edited: %v", err)
	}
	editToTestDuration := time.Since(editToTestStart)
	if editToTestDuration > 5*time.Second {
		t.Fatalf("edit-to-test took %s, expected under 5s", editToTestDuration)
	}

	// j. Run ghega channel diff and verify it detects the change.
	// Diff is performed before deploying the edited version so that it
	// meaningfully detects the pending change.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err = runChannelDiffWithStore([]string{channelPath}, store)
	w.Close()
	os.Stdout = oldStdout
	if err != nil {
		t.Fatalf("diff edited: %v", err)
	}
	diffOut, _ := io.ReadAll(r)
	if !strings.Contains(string(diffOut), "changes detected") {
		t.Fatalf("expected diff to detect changes, got: %s", string(diffOut))
	}

	// i. Deploy the edited channel.
	if err := runChannelDeployWithStore([]string{channelPath}, store); err != nil {
		t.Fatalf("deploy edited: %v", err)
	}

	// k. Run ghega channel rollback and verify it reverts.
	if err := runChannelRollbackWithStore([]string{"e2e-demo", originalRecord.Hash}, store); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	// Verify rollback audit was recorded.
	audit, err := store.ListDeploymentAudit(ctx, "e2e-demo")
	if err != nil {
		t.Fatalf("list deployment audit: %v", err)
	}
	foundRollback := false
	for _, a := range audit {
		if a.Action == "rollback" && a.Hash == originalRecord.Hash {
			foundRollback = true
			break
		}
	}
	if !foundRollback {
		t.Fatal("expected rollback audit entry")
	}

	// l. Verify the rolled-back channel matches the original.
	rolledBackRecord, err := store.GetChannelRevision(ctx, "e2e-demo", originalRecord.Hash)
	if err != nil {
		t.Fatalf("get rolled-back revision: %v", err)
	}
	if string(rolledBackRecord.YAML) != string(originalData) {
		t.Error("rolled-back YAML does not match original")
	}
}
