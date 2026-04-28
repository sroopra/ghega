package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWatchScanAndUpdate(t *testing.T) {
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: watch-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	mtimes := make(map[string]time.Time)
	if err := scanAndUpdate(dir, mtimes); err != nil {
		t.Fatalf("scanAndUpdate: %v", err)
	}

	if len(mtimes) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(mtimes))
	}
	if _, ok := mtimes[chPath]; !ok {
		t.Errorf("expected %s in mtimes", chPath)
	}
}

func TestWatchFindChanged(t *testing.T) {
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: watch-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	mtimes := make(map[string]time.Time)
	if err := scanAndUpdate(dir, mtimes); err != nil {
		t.Fatalf("scanAndUpdate: %v", err)
	}

	// No changes yet.
	changed, err := findChanged(dir, mtimes)
	if err != nil {
		t.Fatalf("findChanged: %v", err)
	}
	if len(changed) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changed))
	}

	// Modify the file.
	time.Sleep(10 * time.Millisecond)
	chYAML2 := `name: watch-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
  - source: MSH-9.1
    target: message_type
`
	if err := os.WriteFile(chPath, []byte(chYAML2), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	changed, err = findChanged(dir, mtimes)
	if err != nil {
		t.Fatalf("findChanged: %v", err)
	}
	if len(changed) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changed))
	}
	if changed[0] != chPath {
		t.Errorf("expected %s, got %s", chPath, changed[0])
	}
}

func TestWatchRunValidation(t *testing.T) {
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: watch-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
tests:
  - name: watch-test
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

	watchNoColor = true
	defer func() { watchNoColor = false }()

	runWatchValidation(chPath)

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	s := string(out)
	if !strings.Contains(s, "PASS validation") {
		t.Errorf("expected PASS validation, got:\n%s", s)
	}
	if !strings.Contains(s, "PASS watch-test") {
		t.Errorf("expected PASS watch-test, got:\n%s", s)
	}
	if !strings.Contains(s, "PASS all tests") {
		t.Errorf("expected PASS all tests, got:\n%s", s)
	}
}

func TestWatchRunValidation_Fail(t *testing.T) {
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: watch-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
tests:
  - name: watch-test
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
    expected:
      patient_mrn: WRONG
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	watchNoColor = true
	defer func() { watchNoColor = false }()

	runWatchValidation(chPath)

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	s := string(out)
	if !strings.Contains(s, "PASS validation") {
		t.Errorf("expected PASS validation, got:\n%s", s)
	}
	if !strings.Contains(s, "FAIL watch-test") {
		t.Errorf("expected FAIL watch-test, got:\n%s", s)
	}
}
