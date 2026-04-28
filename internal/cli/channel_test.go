package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sroopra/ghega/pkg/channel"
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

func TestChannelRollback_WithHash(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	chPath := filepath.Join(dir, "channel.yaml")

	chYAML1 := `name: adt-a01
description: version 1
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
`
	if err := os.WriteFile(chPath, []byte(chYAML1), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	t.Setenv("GHEGA_DATABASE_URL", dbPath)

	if err := runChannelDeploy([]string{chPath}); err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	chYAML2 := `name: adt-a01
description: version 2
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
`
	if err := os.WriteFile(chPath, []byte(chYAML2), 0644); err != nil {
		t.Fatalf("write channel v2: %v", err)
	}

	if err := runChannelDeploy([]string{chPath}); err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	ch1, valErrs := channel.ValidateYAML([]byte(chYAML1))
	if len(valErrs) > 0 {
		t.Fatalf("validate yaml: %v", valErrs)
	}
	hash1, err := channel.HashChannel(ch1)
	if err != nil {
		t.Fatalf("hash channel: %v", err)
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = runChannelRollback([]string{"adt-a01", "--to", hash1})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("rollback with hash: %v", err)
	}

	expected := fmt.Sprintf("Rolled back adt-a01 to hash %s", hash1)
	if !strings.Contains(string(out), expected) {
		t.Errorf("expected %q in output, got:\n%s", expected, string(out))
	}
}

func TestChannelRollback_Auto(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	chPath := filepath.Join(dir, "channel.yaml")

	chYAML1 := `name: adt-a01
description: version 1
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
`
	chYAML2 := `name: adt-a01
description: version 2
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
`

	if err := os.WriteFile(chPath, []byte(chYAML1), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	t.Setenv("GHEGA_DATABASE_URL", dbPath)

	if err := runChannelDeploy([]string{chPath}); err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	if err := os.WriteFile(chPath, []byte(chYAML2), 0644); err != nil {
		t.Fatalf("write channel v2: %v", err)
	}

	if err := runChannelDeploy([]string{chPath}); err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	ch1, valErrs := channel.ValidateYAML([]byte(chYAML1))
	if len(valErrs) > 0 {
		t.Fatalf("validate yaml: %v", valErrs)
	}
	hash1, err := channel.HashChannel(ch1)
	if err != nil {
		t.Fatalf("hash channel: %v", err)
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = runChannelRollback([]string{"adt-a01"})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("auto rollback: %v", err)
	}

	expected := fmt.Sprintf("Rolled back adt-a01 to hash %s", hash1)
	if !strings.Contains(string(out), expected) {
		t.Errorf("expected %q in output, got:\n%s", expected, string(out))
	}
}
