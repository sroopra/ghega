package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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

// Forbidden JS engine imports that must not appear in go.mod or .go source files.
var forbiddenImports = []string{
	"github.com/dop251/goja",
	"github.com/robertkrimen/otto",
	"github.com/duke-git/lancet/v2/javascript",
	"github.com/traefik/yaegi",
	"github.com/containous/yaegi",
	"rogchap.com/v8go",
	"github.com/augustoroman/v8",
	"github.com/ry/v8worker",
	"github.com/andybalholm/gojs",
}

// Forbidden file extensions under internal/, pkg/, cmd/.
var forbiddenExts = []string{".java", ".js", ".ts", ".tsx"}

func TestNoForbiddenImportsInGoMod(t *testing.T) {
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}
	content := string(data)
	for _, imp := range forbiddenImports {
		if strings.Contains(content, imp) {
			t.Errorf("go.mod contains forbidden import: %s", imp)
		}
	}
}

func TestNoForbiddenImportsInSource(t *testing.T) {
	root := repoRoot(t)
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		for _, imp := range forbiddenImports {
			if strings.Contains(content, imp) {
				t.Errorf("%s contains forbidden import: %s", path, imp)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
}

func TestNoForbiddenFileExtensions(t *testing.T) {
	root := repoRoot(t)
	dirs := []string{
		filepath.Join(root, "internal"),
		filepath.Join(root, "pkg"),
		filepath.Join(root, "cmd"),
	}

	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			for _, ext := range forbiddenExts {
				if strings.HasSuffix(path, ext) {
					t.Errorf("forbidden file extension found: %s", path)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk failed for %s: %v", dir, err)
		}
	}
}

func TestRuntimeCheckScriptExistsAndIsExecutable(t *testing.T) {
	root := repoRoot(t)
	script := filepath.Join(root, "scripts", "test-runtime-no-java-js.sh")
	info, err := os.Stat(script)
	if err != nil {
		t.Fatalf("script not found: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("script is not executable: %s", script)
	}
}

func TestRuntimeCheckScriptRuns(t *testing.T) {
	root := repoRoot(t)
	script := filepath.Join(root, "scripts", "test-runtime-no-java-js.sh")
	cmd := exec.Command("bash", script)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("runtime check script failed: %v\noutput:\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "All runtime boundary checks passed") {
		t.Errorf("expected success message in script output, got:\n%s", string(out))
	}
}
