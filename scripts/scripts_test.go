package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoCareMeldReferences(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("failed to determine repo root: %v", err)
	}

	script := filepath.Join(repoRoot, "scripts", "check-branding.sh")
	cmd := exec.Command("bash", script)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("branding check failed: %v\noutput: %s", err, out)
	}

	s := string(out)
	if !strings.Contains(s, "OK") {
		t.Errorf("expected branding check to pass, got:\n%s", s)
	}
}

func TestNoCareMeldInSourceFiles(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("failed to determine repo root: %v", err)
	}

	// Walk the repo and fail if any source file contains CareMeld/caremeld.
	// We exclude .git, plan files, test files, and the check script itself.
	err = filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip broken symlinks or permission errors.
			return nil
		}
		rel, _ := filepath.Rel(repoRoot, path)

		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip historical plan files, test files, and the check script.
		if strings.Contains(rel, "plan") && strings.HasSuffix(rel, ".md") {
			return nil
		}
		if strings.HasSuffix(rel, "_test.go") {
			return nil
		}
		if filepath.Base(rel) == "check-branding.sh" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := strings.ToLower(string(data))
		if strings.Contains(content, "caremeld") {
			t.Errorf("found 'caremeld' reference in %s", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
}
