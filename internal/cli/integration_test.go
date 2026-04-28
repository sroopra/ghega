package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sroopra/ghega/internal/cli/generate"
	"github.com/sroopra/ghega/pkg/channel"
	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestPhase3EndToEndWorkflow(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "channel")

	// a. Generate a channel to a temp directory.
	if err := generate.RunChannelGenerate([]string{"mllp-to-http", "--name", "e2e-channel", "--message-type", "ADT^A01", "--out", outDir}); err != nil {
		t.Fatalf("generate channel: %v", err)
	}

	channelPath := filepath.Join(outDir, "channel.yaml")

	// b. Validate it.
	data, err := os.ReadFile(channelPath)
	if err != nil {
		t.Fatalf("read channel.yaml: %v", err)
	}
	ch, valErrs := channel.ValidateYAML(data)
	if ch == nil {
		t.Fatal("expected channel to parse")
	}
	if len(valErrs) > 0 {
		for _, e := range valErrs {
			t.Errorf("validation error: %s: %s", e.Field, e.Message)
		}
		t.Fatal("channel validation failed")
	}
	policyErrs := channel.ValidatePolicies(ch)
	if len(policyErrs) > 0 {
		for _, e := range policyErrs {
			t.Errorf("policy validation error: %s: %s", e.Field, e.Message)
		}
		t.Fatal("policy validation failed")
	}

	// c. Test it.
	fixtures, err := channel.LoadTestFixtures(channelPath, ch.Tests)
	if err != nil {
		t.Fatalf("load test fixtures: %v", err)
	}
	for _, fixture := range fixtures {
		result, err := channel.RunTest(fixture, ch.Mappings)
		if err != nil {
			t.Fatalf("run test %q: %v", fixture.Name, err)
		}
		if !result.Passed {
			t.Fatalf("test %q failed: %s", fixture.Name, strings.Join(result.Errors, "; "))
		}
	}

	// d. Deploy it using an in-memory channel store.
	store := channelstore.NewInMemoryStore()
	deployResult, err := channel.Deploy(ctx, store, ch, data)
	if err != nil {
		t.Fatalf("deploy channel: %v", err)
	}

	// e. Verify deployment result.
	if deployResult.Name != "e2e-channel" {
		t.Errorf("expected name e2e-channel, got %s", deployResult.Name)
	}
	if deployResult.NoOp {
		t.Fatal("expected first deploy not to be no-op")
	}
	if deployResult.Hash == "" {
		t.Fatal("expected non-empty hash")
	}
	originalHash := deployResult.Hash

	// f. Edit the channel YAML (change a mapping target).
	editStart := time.Now()
	yamlContent := string(data)
	yamlContent = strings.Replace(yamlContent, "target: patient_mrn", "target: patient_id", 1)
	yamlContent = strings.Replace(yamlContent, "patient_mrn: SYNTHETIC_MRN_123456", "patient_id: SYNTHETIC_MRN_123456", 1)
	if err := os.WriteFile(channelPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write edited channel: %v", err)
	}

	// g. Validate the edited channel.
	editedData, err := os.ReadFile(channelPath)
	if err != nil {
		t.Fatalf("read edited channel: %v", err)
	}
	editedCh, valErrs := channel.ValidateYAML(editedData)
	if editedCh == nil {
		t.Fatal("expected edited channel to parse")
	}
	if len(valErrs) > 0 {
		for _, e := range valErrs {
			t.Errorf("validation error: %s: %s", e.Field, e.Message)
		}
		t.Fatal("edited channel validation failed")
	}

	// h. Test the edited channel.
	editedFixtures, err := channel.LoadTestFixtures(channelPath, editedCh.Tests)
	if err != nil {
		t.Fatalf("load edited test fixtures: %v", err)
	}
	for _, fixture := range editedFixtures {
		result, err := channel.RunTest(fixture, editedCh.Mappings)
		if err != nil {
			t.Fatalf("run edited test %q: %v", fixture.Name, err)
		}
		if !result.Passed {
			t.Fatalf("edited test %q failed: %s", fixture.Name, strings.Join(result.Errors, "; "))
		}
	}
	editToTestDuration := time.Since(editStart)
	if editToTestDuration > 5*time.Second {
		t.Fatalf("edit-to-test took %s, expected under 5s", editToTestDuration)
	}

	// j. Run diff and verify it detects the change.
	diffResult, err := channel.DiffLocal(ctx, store, editedCh)
	if err != nil {
		t.Fatalf("diff local: %v", err)
	}
	if diffResult.Identical {
		t.Fatal("expected diff to detect change")
	}
	if !diffResult.Deployed {
		t.Fatal("expected channel to be marked as deployed")
	}

	// i. Deploy the edited channel.
	editedDeployResult, err := channel.Deploy(ctx, store, editedCh, editedData)
	if err != nil {
		t.Fatalf("deploy edited channel: %v", err)
	}
	if editedDeployResult.NoOp {
		t.Fatal("expected edited deploy not to be no-op")
	}
	if editedDeployResult.Hash == originalHash {
		t.Fatal("expected edited hash to differ from original")
	}

	// k. Run rollback and verify it reverts.
	if err := channel.Rollback(ctx, store, "e2e-channel", originalHash); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	// Verify the current channel after rollback matches the original.
	rolledBackRecord, err := store.GetChannel(ctx, "e2e-channel")
	if err != nil {
		t.Fatalf("get rolled-back channel: %v", err)
	}
	rolledBackCh, valErrs := channel.ValidateYAML(rolledBackRecord.YAML)
	if len(valErrs) > 0 {
		for _, e := range valErrs {
			t.Errorf("validation error: %s: %s", e.Field, e.Message)
		}
		t.Fatal("rolled-back channel validation failed")
	}

	// l. Verify the rolled-back channel matches the original via hash.
	rolledBackHash, err := channel.HashChannel(rolledBackCh)
	if err != nil {
		t.Fatalf("hash rolled-back channel: %v", err)
	}
	if rolledBackHash != originalHash {
		t.Fatalf("rolled-back hash %s does not match original %s", rolledBackHash, originalHash)
	}
}
