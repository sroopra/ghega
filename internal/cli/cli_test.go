package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func repoRoot(t *testing.T) string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (go.mod)")
		}
		dir = parent
	}
}

func TestVersionOutput(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/ghega", "version")
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version command failed: %v\noutput: %s", err, out)
	}

	s := string(out)
	if !strings.Contains(s, "Ghega version") {
		t.Errorf("expected output to contain 'Ghega version', got:\n%s", s)
	}
	if !strings.Contains(s, "commit:") {
		t.Errorf("expected output to contain 'commit:', got:\n%s", s)
	}
}

func TestHealthzEndpoint(t *testing.T) {
	port := "18080"
	os.Setenv("GHEGA_PORT", port)
	defer os.Unsetenv("GHEGA_PORT")

	// Start server in background
	go func() {
		_ = runServe([]string{"-port", port})
	}()

	// Wait for server to be ready
	url := fmt.Sprintf("http://localhost:%s/healthz", port)
	var resp *http.Response
	var err error
	for i := 0; i < 50; i++ {
		resp, err = http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("healthz request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"status":"ok"`) {
		t.Errorf("expected body to contain {\"status\":\"ok\"}, got: %s", body)
	}
}

func TestChannelValidateValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	validYAML := `name: test-channel
source:
  type: mllp
destination:
  type: http
mappings:
  - source: PID-3.1
    target: patient_mrn
`
	path := filepath.Join(tmpDir, "channel.yaml")
	if err := os.WriteFile(path, []byte(validYAML), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := runChannelValidate([]string{path})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(string(out), "channel is valid") {
		t.Errorf("expected stdout to contain 'channel is valid', got: %s", string(out))
	}
}

func TestChannelValidateInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	invalidYAML := `name: INVALID_NAME
source:
  type: unknown
destination:
  type: http
mappings: []
`
	path := filepath.Join(tmpDir, "channel.yaml")
	if err := os.WriteFile(path, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	err := runChannelValidate([]string{path})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "name:") {
		t.Errorf("expected error to contain name validation error, got: %s", err.Error())
	}
}

func TestChannelValidateMissingFile(t *testing.T) {
	err := runChannelValidate([]string{"/tmp/nonexistent-channel.yaml"})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "error reading file") {
		t.Errorf("expected error to contain 'error reading file', got: %s", err.Error())
	}
}


