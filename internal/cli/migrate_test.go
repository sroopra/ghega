package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateMirth_Success(t *testing.T) {
	tmpDir := t.TempDir()
	exportDir := filepath.Join("testdata", "mirth-export")
	outDir := filepath.Join(tmpDir, "migrated")

	err := runMigrateMirth([]string{"--out", outDir, exportDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify summary report exists.
	summaryPath := filepath.Join(outDir, "migration-report.yaml")
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Fatalf("expected summary report at %s", summaryPath)
	}

	// Verify per-channel directories and files.
	expectedChannels := []string{"adt-a01-feed", "script-channel", "unsupported-source"}
	for _, ch := range expectedChannels {
		chDir := filepath.Join(outDir, ch)
		for _, file := range []string{"channel.yaml", "rewrite-tasks.yaml", "migration-report.yaml"} {
			p := filepath.Join(chDir, file)
			if _, err := os.Stat(p); os.IsNotExist(err) {
				t.Fatalf("expected %s for channel %s", file, ch)
			}
		}
	}
}

func TestMigrateMirth_MissingOut(t *testing.T) {
	exportDir := filepath.Join("testdata", "mirth-export")
	err := runMigrateMirth([]string{exportDir})
	if err == nil {
		t.Fatal("expected error for missing --out")
	}
	if !strings.Contains(err.Error(), "--out is required") {
		t.Errorf("expected '--out is required' error, got: %s", err.Error())
	}
}

func TestMigrateMirth_MissingExportDir(t *testing.T) {
	err := runMigrateMirth([]string{"--out", "/tmp/migrated"})
	if err == nil {
		t.Fatal("expected error for missing export dir")
	}
	if !strings.Contains(err.Error(), "usage:") {
		t.Errorf("expected usage error, got: %s", err.Error())
	}
}

func TestMigrateMirth_ReportsHaveExpectedStatuses(t *testing.T) {
	tmpDir := t.TempDir()
	exportDir := filepath.Join("testdata", "mirth-export")
	outDir := filepath.Join(tmpDir, "migrated")

	err := runMigrateMirth([]string{"--out", outDir, exportDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// ADT channel has no scripts -> auto-converted.
	adtReport, err := os.ReadFile(filepath.Join(outDir, "adt-a01-feed", "migration-report.yaml"))
	if err != nil {
		t.Fatalf("read adt report: %v", err)
	}
	if !strings.Contains(string(adtReport), "status: auto-converted") {
		t.Errorf("expected adt-a01-feed to be auto-converted, got:\n%s", string(adtReport))
	}

	// Script channel has conditional logic -> needs-rewrite.
	scriptReport, err := os.ReadFile(filepath.Join(outDir, "script-channel", "migration-report.yaml"))
	if err != nil {
		t.Fatalf("read script report: %v", err)
	}
	if !strings.Contains(string(scriptReport), "status: needs-rewrite") {
		t.Errorf("expected script-channel to need rewrite, got:\n%s", string(scriptReport))
	}

	// Unsupported channel has E4X -> unsupported.
	unsupportedReport, err := os.ReadFile(filepath.Join(outDir, "unsupported-source", "migration-report.yaml"))
	if err != nil {
		t.Fatalf("read unsupported report: %v", err)
	}
	if !strings.Contains(string(unsupportedReport), "status: unsupported") {
		t.Errorf("expected unsupported-source to be unsupported, got:\n%s", string(unsupportedReport))
	}
}

func TestMigrateMirth_ChannelYAMLHasMappings(t *testing.T) {
	tmpDir := t.TempDir()
	exportDir := filepath.Join("testdata", "mirth-export")
	outDir := filepath.Join(tmpDir, "migrated")

	err := runMigrateMirth([]string{"--out", outDir, exportDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Script channel should have one auto-converted mapping (static assignment)
	// and rewrite tasks for the conditional.
	chYAML, err := os.ReadFile(filepath.Join(outDir, "script-channel", "channel.yaml"))
	if err != nil {
		t.Fatalf("read channel.yaml: %v", err)
	}
	if !strings.Contains(string(chYAML), "mappings:") {
		t.Errorf("expected channel.yaml to contain mappings, got:\n%s", string(chYAML))
	}

	rtYAML, err := os.ReadFile(filepath.Join(outDir, "script-channel", "rewrite-tasks.yaml"))
	if err != nil {
		t.Fatalf("read rewrite-tasks.yaml: %v", err)
	}
	if !strings.Contains(string(rtYAML), "severity:") {
		t.Errorf("expected rewrite-tasks.yaml to contain tasks, got:\n%s", string(rtYAML))
	}
}
