package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"github.com/sroopra/ghega/pkg/channel"
	"github.com/sroopra/ghega/pkg/migration"
)

func TestMigrateMirth_MissingArgs(t *testing.T) {
	err := runMigrateMirth([]string{})
	if err == nil {
		t.Fatal("expected error for missing args")
	}
	if !strings.Contains(err.Error(), "usage:") {
		t.Errorf("expected usage error, got: %v", err)
	}

	err = runMigrateMirth([]string{"./some-dir"})
	if err == nil {
		t.Fatal("expected error for missing --out")
	}
	if !strings.Contains(err.Error(), "usage:") {
		t.Errorf("expected usage error, got: %v", err)
	}
}

func TestMigrateMirth_ExportDirNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	err := runMigrateMirth([]string{"/nonexistent/export/dir", "--out", filepath.Join(tmpDir, "out")})
	if err == nil {
		t.Fatal("expected error for missing export dir")
	}
	if !strings.Contains(err.Error(), "read export directory") {
		t.Errorf("expected read export directory error, got: %v", err)
	}
}

func TestMigrateMirth_Success(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	testDir := filepath.Join(filepath.Dir(file), "testdata", "mirth-export")
	outDir := t.TempDir()

	err := runMigrateMirth([]string{testDir, "--out", outDir})
	if err != nil {
		t.Fatalf("migrate mirth failed: %v", err)
	}

	// Verify summary report exists.
	summaryPath := filepath.Join(outDir, "migration-report.yaml")
	if _, err := os.Stat(summaryPath); err != nil {
		t.Fatalf("summary report missing: %v", err)
	}

	var summary migration.SummaryReport
	summaryData, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read summary report: %v", err)
	}
	if err := yaml.Unmarshal(summaryData, &summary); err != nil {
		t.Fatalf("unmarshal summary report: %v", err)
	}

	if summary.TotalChannels != 3 {
		t.Errorf("expected 3 channels, got %d", summary.TotalChannels)
	}

	// Verify per-channel directories.
	channels := []struct {
		name           string
		expectedStatus migration.MigrationStatus
	}{
		{"adt-a01-feed", migration.StatusAutoConverted},
		{"script-channel", migration.StatusNeedsRewrite},
		{"unsupported-source", migration.StatusUnsupported},
	}

	for _, ch := range channels {
		chDir := filepath.Join(outDir, ch.name)
		if _, err := os.Stat(chDir); err != nil {
			t.Errorf("channel dir %s missing: %v", ch.name, err)
			continue
		}

		// Check channel.yaml.
		chPath := filepath.Join(chDir, "channel.yaml")
		chData, err := os.ReadFile(chPath)
		if err != nil {
			t.Errorf("channel.yaml missing for %s: %v", ch.name, err)
			continue
		}
		var channelDef channel.Channel
		if err := yaml.Unmarshal(chData, &channelDef); err != nil {
			t.Errorf("unmarshal channel.yaml for %s: %v", ch.name, err)
			continue
		}
		if channelDef.Name != ch.name {
			t.Errorf("channel name mismatch for %s: got %s", ch.name, channelDef.Name)
		}

		// Check rewrite-tasks.yaml.
		rtPath := filepath.Join(chDir, "rewrite-tasks.yaml")
		if _, err := os.Stat(rtPath); err != nil {
			t.Errorf("rewrite-tasks.yaml missing for %s: %v", ch.name, err)
		}

		// Check migration-report.yaml.
		rptPath := filepath.Join(chDir, "migration-report.yaml")
		rptData, err := os.ReadFile(rptPath)
		if err != nil {
			t.Errorf("migration-report.yaml missing for %s: %v", ch.name, err)
			continue
		}
		var rpt migration.ChannelMigrationReport
		if err := yaml.Unmarshal(rptData, &rpt); err != nil {
			t.Errorf("unmarshal migration-report.yaml for %s: %v", ch.name, err)
			continue
		}
		if rpt.Status != ch.expectedStatus {
			t.Errorf("expected status %s for %s, got %s", ch.expectedStatus, ch.name, rpt.Status)
		}
	}
}

func TestMigrateMirth_EmptyDir(t *testing.T) {
	emptyDir := t.TempDir()
	outDir := t.TempDir()

	err := runMigrateMirth([]string{emptyDir, "--out", outDir})
	if err != nil {
		t.Fatalf("migrate mirth on empty dir failed: %v", err)
	}

	summaryPath := filepath.Join(outDir, "migration-report.yaml")
	if _, err := os.Stat(summaryPath); err != nil {
		t.Fatalf("summary report missing: %v", err)
	}

	var summary migration.SummaryReport
	summaryData, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read summary report: %v", err)
	}
	if err := yaml.Unmarshal(summaryData, &summary); err != nil {
		t.Fatalf("unmarshal summary report: %v", err)
	}

	if summary.TotalChannels != 0 {
		t.Errorf("expected 0 channels, got %d", summary.TotalChannels)
	}
}

func TestMigrateMirth_SamplesAndExpectedFlags(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	testDir := filepath.Join(filepath.Dir(file), "testdata", "mirth-export")
	outDir := t.TempDir()
	samplesDir := t.TempDir()
	expectedDir := t.TempDir()

	err := runMigrateMirth([]string{
		testDir,
		"--out", outDir,
		"--samples", samplesDir,
		"--expected", expectedDir,
	})
	if err != nil {
		t.Fatalf("migrate mirth failed: %v", err)
	}
}
