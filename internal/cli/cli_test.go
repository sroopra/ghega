package cli

import (
	"bytes"
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

func TestChannelValidateExits1(t *testing.T) {
	if os.Getenv("BE_TEST_CHANNEL_VALIDATE") == "1" {
		_ = runChannel([]string{"validate", "/tmp/fake-channel.yaml"})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestChannelValidateExits1")
	cmd.Env = append(os.Environ(), "BE_TEST_CHANNEL_VALIDATE=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		// expected
	} else if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, stderr.String())
	} else {
		t.Fatalf("expected exit code 1, got 0 (stderr: %s)", stderr.String())
	}

	if !strings.Contains(stderr.String(), "not yet implemented") {
		t.Errorf("expected stderr to contain 'not yet implemented', got: %s", stderr.String())
	}
}
