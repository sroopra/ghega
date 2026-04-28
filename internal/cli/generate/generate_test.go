package generate

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/sroopra/ghega/pkg/channel"
	"gopkg.in/yaml.v3"
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

func TestRunChannelGenerate_MissingName(t *testing.T) {
	err := RunChannelGenerate([]string{"mllp-to-http", "--out", "/tmp/test-out"})
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
	if !strings.Contains(err.Error(), "--name is required") {
		t.Errorf("expected error to mention --name, got: %v", err)
	}
}

func TestRunChannelGenerate_MissingOut(t *testing.T) {
	err := RunChannelGenerate([]string{"mllp-to-http", "--name", "test-channel"})
	if err == nil {
		t.Fatal("expected error for missing --out")
	}
	if !strings.Contains(err.Error(), "--out is required") {
		t.Errorf("expected error to mention --out, got: %v", err)
	}
}

func TestRunChannelGenerate_UnknownType(t *testing.T) {
	err := RunChannelGenerate([]string{"unknown-type"})
	if err == nil {
		t.Fatal("expected error for unknown channel type")
	}
	if !strings.Contains(err.Error(), "unknown channel type") {
		t.Errorf("expected error to mention unknown channel type, got: %v", err)
	}
}

func TestGenerateMLLPToHTTP_CreatesExpectedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "generated")

	err := RunChannelGenerate([]string{"mllp-to-http", "--name", "demo-channel", "--message-type", "ADT_A01", "--out", outDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFiles := []string{
		"channel.yaml",
		"tests/fixture.yaml",
		"fixtures/sample.hl7",
	}

	for _, rel := range expectedFiles {
		path := filepath.Join(outDir, rel)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", rel)
		}
	}
}

func TestGenerateMLLPToHTTP_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "generated")

	err := RunChannelGenerate([]string{"mllp-to-http", "--name", "demo-channel", "--message-type", "ADT_A01", "--out", outDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	channelPath := filepath.Join(outDir, "channel.yaml")
	data, err := os.ReadFile(channelPath)
	if err != nil {
		t.Fatalf("reading channel.yaml: %v", err)
	}

	var doc map[string]interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Errorf("channel.yaml is not valid YAML: %v", err)
	}

	if doc["name"] != "demo-channel" {
		t.Errorf("expected name 'demo-channel', got %v", doc["name"])
	}

	mappings, ok := doc["mappings"].([]interface{})
	if !ok {
		t.Fatalf("expected mappings to be a list, got %T", doc["mappings"])
	}
	if len(mappings) < 4 {
		t.Errorf("expected at least 4 mappings, got %d", len(mappings))
	}
}

func TestGenerateMLLPToHTTP_PassesValidation(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "generated")

	err := RunChannelGenerate([]string{"mllp-to-http", "--name", "demo-channel", "--message-type", "ADT_A01", "--out", outDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	channelPath := filepath.Join(outDir, "channel.yaml")
	data, err := os.ReadFile(channelPath)
	if err != nil {
		t.Fatalf("reading channel.yaml: %v", err)
	}

	ch, valErrs := channel.ValidateYAML(data)
	if ch != nil {
		valErrs = append(valErrs, channel.ValidatePolicies(ch)...)
	}
	if len(valErrs) > 0 {
		var msgs []string
		for _, e := range valErrs {
			msgs = append(msgs, e.Field+": "+e.Message)
		}
		t.Fatalf("generated channel failed validation: %s", strings.Join(msgs, "; "))
	}
}

func TestGenerateMLLPToHTTP_HasErrorHandlingTest(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "generated")

	err := RunChannelGenerate([]string{"mllp-to-http", "--name", "demo-channel", "--message-type", "ADT_A01", "--out", outDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	channelPath := filepath.Join(outDir, "channel.yaml")
	data, err := os.ReadFile(channelPath)
	if err != nil {
		t.Fatalf("reading channel.yaml: %v", err)
	}

	var ch channel.Channel
	if err := yaml.Unmarshal(data, &ch); err != nil {
		t.Fatalf("parsing channel.yaml: %v", err)
	}

	if len(ch.Tests) < 2 {
		t.Fatalf("expected at least 2 tests, got %d", len(ch.Tests))
	}

	foundExpectError := false
	for _, tt := range ch.Tests {
		if tt.ExpectError {
			foundExpectError = true
			break
		}
	}
	if !foundExpectError {
		t.Fatal("expected at least one test with expectError: true")
	}
}

func TestGenerateMLLPToHTTP_NoPHI(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "generated")

	err := RunChannelGenerate([]string{"mllp-to-http", "--name", "demo-channel", "--message-type", "ADT_A01", "--out", outDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	phiPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),               // SSN
		regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`), // email
		regexp.MustCompile(`(?i)password\s*=\s*\S+`),              // password=
		regexp.MustCompile(`(?i)api_key\s*=\s*\S+`),               // api_key=
		regexp.MustCompile(`(?i)secret\s*=\s*\S+`),                // secret=
		regexp.MustCompile(`(?i)token\s*=\s*\S+`),                 // token=
	}

	err = filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		for _, re := range phiPatterns {
			if re.MatchString(content) {
				t.Errorf("potential secret/PHI pattern found in %s: %s", path, re.String())
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking generated files: %v", err)
	}
}

func TestGenerateMLLPToHTTP_ContainsGhegaHeader(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "generated")

	err := RunChannelGenerate([]string{"mllp-to-http", "--name", "demo-channel", "--out", outDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := []string{"channel.yaml", "tests/fixture.yaml"}
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(outDir, rel))
		if err != nil {
			t.Fatalf("reading %s: %v", rel, err)
		}
		if !strings.Contains(string(data), "Generated by Ghega") {
			t.Errorf("expected %s to contain 'Generated by Ghega' header", rel)
		}
	}
}

func TestGenerateMLLPToHTTP_PassesCLIValidation(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "generated")

	err := RunChannelGenerate([]string{"mllp-to-http", "--name", "demo-channel", "--message-type", "ADT_A01", "--out", outDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	channelPath := filepath.Join(outDir, "channel.yaml")
	cmd := exec.Command("go", "run", "./cmd/ghega", "channel", "validate", channelPath)
	cmd.Dir = repoRoot(t)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		t.Fatalf("ghega channel validate failed: %v (stderr: %s)", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "channel is valid") {
		t.Errorf("expected stdout to contain 'channel is valid', got: %s", stdout.String())
	}
}
