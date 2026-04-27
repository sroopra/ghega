package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestChannelTest_AllPass(t *testing.T) {
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: adt-a01
description: Test channel
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: basic-pass
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
    expected:
      patient_mrn: MRN12345
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runChannelTest([]string{chPath})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("runChannelTest: %v", err)
	}
	if !strings.Contains(string(out), "PASS basic-pass") {
		t.Errorf("expected PASS basic-pass, got:\n%s", string(out))
	}
}

func TestChannelTest_WithFailure(t *testing.T) {
	if os.Getenv("BE_TEST_CHANNEL_TEST_FAIL") == "1" {
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
tests:
  - name: basic-fail
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
    expected:
      patient_mrn: WRONG
`
		if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
			t.Fatalf("write channel: %v", err)
		}
		_ = runChannelTest([]string{chPath})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestChannelTest_WithFailure")
	cmd.Env = append(os.Environ(), "BE_TEST_CHANNEL_TEST_FAIL=1")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()

	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		// expected
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		t.Fatalf("expected exit code 1, got 0")
	}

	out := stdout.String()
	if !strings.Contains(out, "FAIL basic-fail:") {
		t.Errorf("expected FAIL basic-fail, got:\n%s", out)
	}
	if !strings.Contains(out, `expected "patient_mrn" = "WRONG"`) {
		t.Errorf("expected error detail, got:\n%s", out)
	}
}

func TestChannelTest_JUnitOutput(t *testing.T) {
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	junitPath := filepath.Join(dir, "report.xml")
	chYAML := `name: adt-a01
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: junit-test
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
    expected:
      patient_mrn: MRN12345
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	err := runChannelTest([]string{"--junit", junitPath, chPath})
	if err != nil {
		t.Fatalf("runChannelTest: %v", err)
	}

	data, err := os.ReadFile(junitPath)
	if err != nil {
		t.Fatalf("read junit: %v", err)
	}
	if !strings.Contains(string(data), `<testcase name="junit-test"`) {
		t.Errorf("expected junit testcase, got:\n%s", string(data))
	}
}

func TestChannelTest_FileFixture(t *testing.T) {
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	testdataDir := filepath.Join(dir, "testdata")
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	hl7Data := "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
	if err := os.WriteFile(filepath.Join(testdataDir, "sample.hl7"), []byte(hl7Data), 0644); err != nil {
		t.Fatalf("write hl7: %v", err)
	}
	chYAML := `name: adt-a01
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: file-fixture
    input: testdata/sample.hl7
    expected:
      patient_mrn: MRN12345
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := runChannelTest([]string{chPath})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("runChannelTest: %v", err)
	}
	if !strings.Contains(string(out), "PASS file-fixture") {
		t.Errorf("expected PASS file-fixture, got:\n%s", string(out))
	}
}

func TestChannelDeploy(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	origNewStore := newStore
	newStore = func() (channelstore.ChannelStore, error) { return store, nil }
	defer func() { newStore = origNewStore }()

	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: deploy-test
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: deploy-pass
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
    expected:
      patient_mrn: MRN12345
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runChannelDeploy([]string{chPath})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("runChannelDeploy: %v", err)
	}
	if !strings.Contains(string(out), "deployed deploy-test") {
		t.Errorf("expected deploy confirmation, got:\n%s", string(out))
	}

	ctx := context.Background()
	rec, err := store.GetChannel(ctx, "deploy-test")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	if rec.Name != "deploy-test" {
		t.Errorf("Name = %q, want %q", rec.Name, "deploy-test")
	}

	audits, err := store.ListDeploymentAudit(ctx, "deploy-test")
	if err != nil {
		t.Fatalf("ListDeploymentAudit: %v", err)
	}
	if len(audits) != 1 {
		t.Fatalf("expected 1 audit, got %d", len(audits))
	}
	if audits[0].Action != "deploy" {
		t.Errorf("Action = %q, want %q", audits[0].Action, "deploy")
	}
}

func TestChannelDiff_NewChannel(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	origNewStore := newStore
	newStore = func() (channelstore.ChannelStore, error) { return store, nil }
	defer func() { newStore = origNewStore }()

	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: diff-test
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: diff-pass
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
    expected:
      patient_mrn: MRN12345
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runChannelDiff([]string{chPath})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("runChannelDiff: %v", err)
	}
	if !strings.Contains(string(out), "not yet deployed") {
		t.Errorf("expected 'not yet deployed' message, got:\n%s", string(out))
	}
}

func TestChannelDiff_UpToDate(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	origNewStore := newStore
	newStore = func() (channelstore.ChannelStore, error) { return store, nil }
	defer func() { newStore = origNewStore }()

	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: diff-uptodate
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: diff-pass
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
    expected:
      patient_mrn: MRN12345
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	// Deploy first.
	if err := runChannelDeploy([]string{chPath}); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runChannelDiff([]string{chPath})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("runChannelDiff: %v", err)
	}
	if !strings.Contains(string(out), "up to date") {
		t.Errorf("expected 'up to date' message, got:\n%s", string(out))
	}
}

func TestChannelDiff_HasChanges(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	origNewStore := newStore
	newStore = func() (channelstore.ChannelStore, error) { return store, nil }
	defer func() { newStore = origNewStore }()

	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: diff-changes
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: diff-pass
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
    expected:
      patient_mrn: MRN12345
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	// Deploy first.
	if err := runChannelDeploy([]string{chPath}); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	// Modify the file.
	modifiedYAML := strings.Replace(chYAML, "MRN12345", "MRN99999", -1)
	if err := os.WriteFile(chPath, []byte(modifiedYAML), 0644); err != nil {
		t.Fatalf("write modified channel: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runChannelDiff([]string{chPath})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("runChannelDiff: %v", err)
	}
	if !strings.Contains(string(out), "has changes") {
		t.Errorf("expected 'has changes' message, got:\n%s", string(out))
	}
}

func TestChannelRollback(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	origNewStore := newStore
	newStore = func() (channelstore.ChannelStore, error) { return store, nil }
	defer func() { newStore = origNewStore }()

	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: rollback-test
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: rollback-pass
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
    expected:
      patient_mrn: MRN12345
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	// Deploy first to get a hash in the store.
	if err := runChannelDeploy([]string{chPath}); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	ctx := context.Background()
	rec, err := store.GetChannel(ctx, "rollback-test")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	hash := rec.Hash

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runChannelRollback([]string{"rollback-test", hash})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("runChannelRollback: %v", err)
	}
	if !strings.Contains(string(out), "rolled back rollback-test") {
		t.Errorf("expected rollback confirmation, got:\n%s", string(out))
	}

	audits, err := store.ListDeploymentAudit(ctx, "rollback-test")
	if err != nil {
		t.Fatalf("ListDeploymentAudit: %v", err)
	}
	if len(audits) != 2 {
		t.Fatalf("expected 2 audits, got %d", len(audits))
	}
	if audits[0].Action != "rollback" {
		t.Errorf("Action = %q, want %q", audits[0].Action, "rollback")
	}
}

func TestChannelRollback_InvalidHash(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	origNewStore := newStore
	newStore = func() (channelstore.ChannelStore, error) { return store, nil }
	defer func() { newStore = origNewStore }()

	if err := runChannelRollback([]string{"missing-channel", "badhash"}); err == nil {
		t.Fatal("expected error for invalid hash")
	}
}

func TestChannelDeploy_InvalidChannel(t *testing.T) {
	if os.Getenv("BE_TEST_CHANNEL_DEPLOY_INVALID") == "1" {
		dir := t.TempDir()
		chPath := filepath.Join(dir, "channel.yaml")
		chYAML := `name: INVALID_NAME
source:
  type: unknown
destination:
  type: http
mappings: []
`
		if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
			t.Fatalf("write channel: %v", err)
		}
		_ = runChannelDeploy([]string{chPath})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestChannelDeploy_InvalidChannel")
	cmd.Env = append(os.Environ(), "BE_TEST_CHANNEL_DEPLOY_INVALID=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		// expected
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		t.Fatalf("expected exit code 1, got 0")
	}

	if !strings.Contains(stderr.String(), "name:") {
		t.Errorf("expected name validation error, got:\n%s", stderr.String())
	}
}
