package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	fnErr := fn()

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	return string(out), fnErr
}

func TestChannelDeploy_CLI(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("GHEGA_DATABASE_URL", dbPath)

	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")
	yaml := `name: cli-test-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	out, err := captureStdout(t, func() error {
		return runChannelDeploy([]string{path})
	})
	if err != nil {
		t.Fatalf("runChannelDeploy failed: %v", err)
	}
	if !strings.Contains(out, "Deployed cli-test-ch revision 1 hash") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestChannelDiff_CLI_NoChanges(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("GHEGA_DATABASE_URL", dbPath)

	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")
	yaml := `name: cli-test-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Deploy first
	if _, err := captureStdout(t, func() error {
		return runChannelDeploy([]string{path})
	}); err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	out, err := captureStdout(t, func() error {
		return runChannelDiff([]string{path})
	})
	if err != nil {
		t.Fatalf("runChannelDiff failed: %v", err)
	}
	if !strings.Contains(out, "No changes") {
		t.Errorf("expected 'No changes', got: %q", out)
	}
}

func TestChannelDiff_CLI_HasChanges(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("GHEGA_DATABASE_URL", dbPath)

	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")
	yaml1 := `name: cli-test-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	if err := os.WriteFile(path, []byte(yaml1), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := captureStdout(t, func() error {
		return runChannelDeploy([]string{path})
	}); err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	yaml2 := `name: cli-test-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
  - source: PID-5.1
    target: last_name
`
	if err := os.WriteFile(path, []byte(yaml2), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	out, err := captureStdout(t, func() error {
		return runChannelDiff([]string{path})
	})
	if err != nil {
		t.Fatalf("runChannelDiff failed: %v", err)
	}
	if !strings.Contains(out, "has changes") {
		t.Errorf("expected 'has changes', got: %q", out)
	}
}

func TestChannelRollback_CLI(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("GHEGA_DATABASE_URL", dbPath)

	dir := t.TempDir()
	path := filepath.Join(dir, "channel.yaml")
	yaml1 := `name: cli-test-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	if err := os.WriteFile(path, []byte(yaml1), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := captureStdout(t, func() error {
		return runChannelDeploy([]string{path})
	}); err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	yaml2 := `name: cli-test-ch
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
  - source: PID-5.1
    target: last_name
`
	if err := os.WriteFile(path, []byte(yaml2), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := captureStdout(t, func() error {
		return runChannelDeploy([]string{path})
	}); err != nil {
		t.Fatalf("deploy second time failed: %v", err)
	}

	out, err := captureStdout(t, func() error {
		return runChannelRollback([]string{"cli-test-ch"})
	})
	if err != nil {
		t.Fatalf("runChannelRollback failed: %v", err)
	}
	if !strings.Contains(out, "Rolled back channel cli-test-ch") {
		t.Errorf("unexpected output: %q", out)
	}
}
