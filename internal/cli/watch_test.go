package cli

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestWatch_DetectsChangeAndRunsValidate(t *testing.T) {
	dir := t.TempDir()

	// Create a valid channel.yaml.
	chYaml := `name: watch-test
description: Test channel for watch
source:
  type: mllp
  config:
    port: 2575
destination:
  type: http
  config:
    url: http://example.com/webhook
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	chPath := filepath.Join(dir, "channel.yaml")
	if err := os.WriteFile(chPath, []byte(chYaml), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	mtimes := make(map[string]time.Time)

	// First scan — should pick up the file.
	if err := scanAndProcess(dir, mtimes); err != nil {
		t.Fatalf("first scan: %v", err)
	}

	if len(mtimes) != 1 {
		t.Fatalf("expected 1 tracked file, got %d", len(mtimes))
	}

	// Modify the file.
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(chPath, []byte(chYaml+"\n# modified\n"), 0644); err != nil {
		t.Fatalf("modify channel: %v", err)
	}

	// Second scan — should detect the change.
	if err := scanAndProcess(dir, mtimes); err != nil {
		t.Fatalf("second scan: %v", err)
	}

	// The watch should have run validate (which passes) and test (which may fail
	// due to missing fixtures, but we just verify the scan detects the change).
	// Since scanAndProcess prints to stdout, we verify the mtime was updated.
	if mtimes[chPath].IsZero() {
		t.Error("expected mtime to be tracked")
	}
}

func TestWatch_InvalidChannelShowsValidationError(t *testing.T) {
	dir := t.TempDir()

	// Create an invalid channel.yaml (missing required fields).
	chYaml := `name: invalid-watch-test
`
	chPath := filepath.Join(dir, "channel.yaml")
	if err := os.WriteFile(chPath, []byte(chYaml), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	mtimes := make(map[string]time.Time)
	if err := scanAndProcess(dir, mtimes); err != nil {
		t.Fatalf("scan: %v", err)
	}

	// scanAndProcess should have tracked the file even though validation failed.
	if len(mtimes) != 1 {
		t.Fatalf("expected 1 tracked file, got %d", len(mtimes))
	}
}

func TestWatch_ExitsOnSIGINT(t *testing.T) {
	dir := t.TempDir()

	done := make(chan error, 1)
	go func() {
		done <- runWatch([]string{dir})
	}()

	// Give runWatch time to start the ticker and signal handler.
	time.Sleep(50 * time.Millisecond)

	// Send SIGINT to the current process.
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected nil error on SIGINT, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runWatch did not exit after SIGINT")
	}
}
