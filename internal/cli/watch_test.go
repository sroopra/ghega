package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && ((s[j] >= '0' && s[j] <= '9') || s[j] == ';') {
				j++
			}
			if j < len(s) {
				i = j // the loop increment will move past the terminating letter
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func TestFindChannelYAMLs(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create channel.yaml files at different depths.
	if err := os.WriteFile(filepath.Join(dir, "channel.yaml"), []byte("a"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "channel.yaml"), []byte("b"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// A file that should not match.
	if err := os.WriteFile(filepath.Join(dir, "other.yaml"), []byte("c"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	files, err := findChannelYAMLs(dir)
	if err != nil {
		t.Fatalf("findChannelYAMLs: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(files), files)
	}
}

func TestScanAndRun_DetectsChange(t *testing.T) {
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: watch-test
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: watch-pass
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
    expected:
      patient_mrn: MRN12345
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	state := make(map[string]time.Time)

	// First scan should populate state but not run tests.
	if err := scanAndRun(dir, state); err != nil {
		t.Fatalf("first scan: %v", err)
	}
	if _, ok := state[chPath]; !ok {
		t.Fatal("expected state to contain channel path after first scan")
	}

	// Modify the file.
	modifiedYAML := strings.Replace(chYAML, "MRN12345", "MRN99999", 1)
	modifiedYAML = strings.Replace(modifiedYAML, "MRN12345", "MRN99999", 1)
	time.Sleep(10 * time.Millisecond) // ensure mtime changes
	if err := os.WriteFile(chPath, []byte(modifiedYAML), 0644); err != nil {
		t.Fatalf("write modified channel: %v", err)
	}

	// Capture stdout.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := scanAndRun(dir, state); err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("second scan: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	output := stripANSI(string(out))

	if !strings.Contains(output, "[change detected]") {
		t.Errorf("expected change detected message, got:\n%s", output)
	}
	if !strings.Contains(output, "PASS watch-pass") {
		t.Errorf("expected PASS watch-pass, got:\n%s", output)
	}
}

func TestValidateAndTest(t *testing.T) {
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: validate-test
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: vt-pass
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

	err := validateAndTest(chPath)

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	output := stripANSI(string(out))

	if err != nil {
		t.Fatalf("validateAndTest: %v", err)
	}
	if !strings.Contains(output, "PASS vt-pass") {
		t.Errorf("expected PASS vt-pass, got:\n%s", output)
	}
}

func TestValidateAndTest_FailingTest(t *testing.T) {
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: validate-test-fail
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: vt-fail
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||WRONG_MRN\r"
    expected:
      patient_mrn: EXPECTED_MRN
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := validateAndTest(chPath)

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	output := stripANSI(string(out))

	if err == nil {
		t.Fatal("expected error when test fails")
	}
	if !strings.Contains(err.Error(), "one or more tests failed") {
		t.Errorf("expected 'one or more tests failed' error, got: %v", err)
	}
	if !strings.Contains(output, "FAIL vt-fail") {
		t.Errorf("expected FAIL vt-fail, got:\n%s", output)
	}
}

func TestValidateAndTest_InvalidChannel(t *testing.T) {
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

	if err := validateAndTest(chPath); err == nil {
		t.Fatal("expected error for invalid channel")
	} else if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected validation failed error, got: %v", err)
	}
}

func TestRunWatch_InvalidDirectory(t *testing.T) {
	if err := runWatch([]string{"/nonexistent/path/12345"}); err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestRunWatch_FileNotDirectory(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "notadir")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := runWatch([]string{f}); err == nil {
		t.Fatal("expected error when path is a file")
	}
}

func TestRunWatch_SigInt(t *testing.T) {
	dir := t.TempDir()
	chPath := filepath.Join(dir, "channel.yaml")
	chYAML := `name: sigint-test
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
    transform: copy
tests:
  - name: sigint-pass
    input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
    expected:
      patient_mrn: MRN12345
`
	if err := os.WriteFile(chPath, []byte(chYAML), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	// Send SIGINT to ourselves after a short delay.
	go func() {
		time.Sleep(200 * time.Millisecond)
		proc, _ := os.FindProcess(os.Getpid())
		if proc != nil {
			_ = proc.Signal(os.Interrupt)
		}
	}()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runWatch([]string{dir})

	w.Close()
	os.Stdout = oldStdout
	_ = r

	if err != nil {
		t.Fatalf("runWatch: %v", err)
	}
}
