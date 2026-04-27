package cli

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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

	var stdout bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runChannelTest([]string{chPath})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	_ = stdout

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
