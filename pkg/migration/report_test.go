package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sroopra/ghega/pkg/channel"
	"github.com/sroopra/ghega/pkg/mapping"
	"github.com/sroopra/ghega/pkg/mirthxml"
	"gopkg.in/yaml.v3"
)

func TestGenerateMigrationReport_AutoConverted(t *testing.T) {
	convResult := &ConversionResult{
		Channel: channel.Channel{
			Name:        "test-channel",
			Description: "Test",
			Source:      channel.Source{Type: "mllp"},
			Destination: channel.Destination{Type: "http"},
		},
		Warnings: []string{},
	}

	report, rewriteTasks, autoMappings := GenerateMigrationReport("test-channel", convResult, nil)

	if report.Status != StatusAutoConverted {
		t.Errorf("expected status %s, got %s", StatusAutoConverted, report.Status)
	}
	if len(report.AutoConverted) != 2 {
		t.Errorf("expected 2 auto-converted items, got %d", len(report.AutoConverted))
	}
	if len(rewriteTasks) != 0 {
		t.Errorf("expected 0 rewrite tasks, got %d", len(rewriteTasks))
	}
	if len(autoMappings) != 0 {
		t.Errorf("expected 0 auto mappings, got %d", len(autoMappings))
	}
}

func TestGenerateMigrationReport_WithMapping(t *testing.T) {
	convResult := &ConversionResult{
		Channel: channel.Channel{
			Name:        "map-channel",
			Source:      channel.Source{Type: "http"},
			Destination: channel.Destination{Type: "file"},
		},
		Warnings: []string{"sample warning"},
	}

	classifications := []ClassificationResult{
		{
			Step: mirthxml.Step{Name: "Set MRN", Script: "msg['PID']['PID.3']['PID.3.1'] = 'UNKNOWN';"},
			Patterns: []ClassifiedPattern{
				{
					Category:    CategoryFieldAssignment,
					Disposition: DispositionAutoConvertible,
					Description: "Simple field assignment from a static value.",
					Mapping: &mapping.Mapping{
						Target:    "PID-3.1",
						Transform: mapping.TransformStatic,
						Value:     "UNKNOWN",
					},
				},
			},
		},
	}

	report, rewriteTasks, autoMappings := GenerateMigrationReport("map-channel", convResult, classifications)

	if report.Status != StatusAutoConverted {
		t.Errorf("expected status %s, got %s", StatusAutoConverted, report.Status)
	}
	if len(autoMappings) != 1 {
		t.Errorf("expected 1 auto mapping, got %d", len(autoMappings))
	}
	if autoMappings[0].Target != "PID-3.1" {
		t.Errorf("expected target PID-3.1, got %s", autoMappings[0].Target)
	}
	if len(rewriteTasks) != 0 {
		t.Errorf("expected 0 rewrite tasks, got %d", len(rewriteTasks))
	}
	if len(report.Warnings) != 1 || report.Warnings[0] != "sample warning" {
		t.Errorf("expected warnings to be preserved, got %v", report.Warnings)
	}
}

func TestGenerateMigrationReport_NeedsRewrite(t *testing.T) {
	convResult := &ConversionResult{
		Channel: channel.Channel{
			Name:        "rewrite-channel",
			Source:      channel.Source{Type: "mllp"},
			Destination: channel.Destination{Type: "http"},
		},
		Warnings: []string{},
	}

	classifications := []ClassificationResult{
		{
			Step: mirthxml.Step{Name: "Conditional", Script: "if (true) { msg['PID']['PID.3']['PID.3.1'] = 'X'; }"},
			Patterns: []ClassifiedPattern{
				{
					Category:    CategoryConditional,
					Disposition: DispositionNeedsRewrite,
					Description: "Conditional logic detected.",
					RewriteTask: &RewriteTask{
						Severity:    "high",
						Description: "Replace conditional with mapping.",
					},
				},
			},
		},
	}

	report, rewriteTasks, autoMappings := GenerateMigrationReport("rewrite-channel", convResult, classifications)

	if report.Status != StatusNeedsRewrite {
		t.Errorf("expected status %s, got %s", StatusNeedsRewrite, report.Status)
	}
	if len(report.NeedsRewrite) != 1 {
		t.Errorf("expected 1 needs-rewrite item, got %d", len(report.NeedsRewrite))
	}
	if len(rewriteTasks) != 1 {
		t.Errorf("expected 1 rewrite task, got %d", len(rewriteTasks))
	}
	if rewriteTasks[0].Category != CategoryConditional {
		t.Errorf("expected category conditional, got %s", rewriteTasks[0].Category)
	}
	if len(autoMappings) != 0 {
		t.Errorf("expected 0 auto mappings, got %d", len(autoMappings))
	}
}

func TestGenerateMigrationReport_Unsupported(t *testing.T) {
	convResult := &ConversionResult{
		Channel: channel.Channel{
			Name:        "unsupported-channel",
			Source:      channel.Source{Type: "mllp"},
			Destination: channel.Destination{Type: "http"},
		},
		Warnings: []string{},
	}

	classifications := []ClassificationResult{
		{
			Step: mirthxml.Step{Name: "E4X", Script: "var x = new XML('<r/>');"},
			Patterns: []ClassifiedPattern{
				{
					Category:    CategoryE4XManipulation,
					Disposition: DispositionUnsupported,
					Description: "E4X/XML manipulation detected.",
					RewriteTask: &RewriteTask{
						Severity:    "high",
						Description: "Rewrite E4X.",
					},
				},
			},
		},
	}

	report, rewriteTasks, autoMappings := GenerateMigrationReport("unsupported-channel", convResult, classifications)

	if report.Status != StatusUnsupported {
		t.Errorf("expected status %s, got %s", StatusUnsupported, report.Status)
	}
	if len(report.Unsupported) != 1 {
		t.Errorf("expected 1 unsupported item, got %d", len(report.Unsupported))
	}
	if len(rewriteTasks) != 1 {
		t.Errorf("expected 1 rewrite task, got %d", len(rewriteTasks))
	}
	if len(autoMappings) != 0 {
		t.Errorf("expected 0 auto mappings, got %d", len(autoMappings))
	}
}

func TestGenerateMigrationReport_Mixed(t *testing.T) {
	convResult := &ConversionResult{
		Channel: channel.Channel{
			Name:        "mixed-channel",
			Source:      channel.Source{Type: "mllp"},
			Destination: channel.Destination{Type: "http"},
		},
		Warnings: []string{},
	}

	classifications := []ClassificationResult{
		{
			Step: mirthxml.Step{Name: "Set MRN", Script: "msg['PID']['PID.3']['PID.3.1'] = 'UNKNOWN';"},
			Patterns: []ClassifiedPattern{
				{
					Category:    CategoryFieldAssignment,
					Disposition: DispositionAutoConvertible,
					Description: "Simple field assignment from a static value.",
					Mapping: &mapping.Mapping{
						Target:    "PID-3.1",
						Transform: mapping.TransformStatic,
						Value:     "UNKNOWN",
					},
				},
			},
		},
		{
			Step: mirthxml.Step{Name: "Loop", Script: "for each (seg in msg.children()) { logger.info(seg); }"},
			Patterns: []ClassifiedPattern{
				{
					Category:    CategoryLoop,
					Disposition: DispositionNeedsRewrite,
					Description: "Loop construct detected.",
					RewriteTask: &RewriteTask{
						Severity:    "high",
						Description: "Replace loop with mapping.",
					},
				},
			},
		},
	}

	report, _, _ := GenerateMigrationReport("mixed-channel", convResult, classifications)

	if report.Status != StatusNeedsRewrite {
		t.Errorf("expected status %s, got %s", StatusNeedsRewrite, report.Status)
	}
	if len(report.AutoConverted) != 3 { // 2 connectors + 1 mapping
		t.Errorf("expected 3 auto-converted items, got %d", len(report.AutoConverted))
	}
	if len(report.NeedsRewrite) != 1 {
		t.Errorf("expected 1 needs-rewrite item, got %d", len(report.NeedsRewrite))
	}
}

func TestWriteChannelOutput(t *testing.T) {
	tmpDir := t.TempDir()

	convResult := &ConversionResult{
		Channel: channel.Channel{
			Name:        "write-test",
			Source:      channel.Source{Type: "mllp"},
			Destination: channel.Destination{Type: "http"},
		},
		Warnings: []string{"warn1"},
	}

	report := &ChannelMigrationReport{
		ChannelName:   "write-test",
		Status:        StatusAutoConverted,
		AutoConverted: []string{"source connector"},
		Warnings:      []string{"warn1"},
	}

	rewriteTasks := []TypedRewriteTask{}
	autoMappings := []mapping.Mapping{{Target: "PID-3.1", Transform: mapping.TransformStatic, Value: "X"}}

	err := WriteChannelOutput(tmpDir, "write-test", convResult, report, rewriteTasks, autoMappings)
	if err != nil {
		t.Fatalf("write channel output: %v", err)
	}

	// Verify channel.yaml exists and contains mapping.
	chPath := filepath.Join(tmpDir, "write-test", "channel.yaml")
	chData, err := os.ReadFile(chPath)
	if err != nil {
		t.Fatalf("read channel.yaml: %v", err)
	}
	var ch channel.Channel
	if err := yaml.Unmarshal(chData, &ch); err != nil {
		t.Fatalf("unmarshal channel.yaml: %v", err)
	}
	if len(ch.Mappings) != 1 {
		t.Errorf("expected 1 mapping in channel.yaml, got %d", len(ch.Mappings))
	}

	// Verify rewrite-tasks.yaml exists.
	rtPath := filepath.Join(tmpDir, "write-test", "rewrite-tasks.yaml")
	if _, err := os.Stat(rtPath); err != nil {
		t.Fatalf("rewrite-tasks.yaml missing: %v", err)
	}

	// Verify migration-report.yaml exists.
	rptPath := filepath.Join(tmpDir, "write-test", "migration-report.yaml")
	rptData, err := os.ReadFile(rptPath)
	if err != nil {
		t.Fatalf("read migration-report.yaml: %v", err)
	}
	var rpt ChannelMigrationReport
	if err := yaml.Unmarshal(rptData, &rpt); err != nil {
		t.Fatalf("unmarshal migration-report.yaml: %v", err)
	}
	if rpt.ChannelName != "write-test" {
		t.Errorf("expected channelName write-test, got %s", rpt.ChannelName)
	}
}

func TestWriteSummaryReport(t *testing.T) {
	tmpDir := t.TempDir()

	reports := []*ChannelMigrationReport{
		{ChannelName: "ch1", Status: StatusAutoConverted},
		{ChannelName: "ch2", Status: StatusNeedsRewrite, Warnings: []string{"w1"}},
		{ChannelName: "ch3", Status: StatusUnsupported},
	}

	err := WriteSummaryReport(tmpDir, reports)
	if err != nil {
		t.Fatalf("write summary report: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "migration-report.yaml"))
	if err != nil {
		t.Fatalf("read summary report: %v", err)
	}

	var summary SummaryReport
	if err := yaml.Unmarshal(data, &summary); err != nil {
		t.Fatalf("unmarshal summary report: %v", err)
	}

	if summary.TotalChannels != 3 {
		t.Errorf("expected total 3, got %d", summary.TotalChannels)
	}
	if summary.AutoConverted != 1 {
		t.Errorf("expected auto-converted 1, got %d", summary.AutoConverted)
	}
	if summary.NeedsRewrite != 1 {
		t.Errorf("expected needs-rewrite 1, got %d", summary.NeedsRewrite)
	}
	if summary.Unsupported != 1 {
		t.Errorf("expected unsupported 1, got %d", summary.Unsupported)
	}
	if len(summary.Channels) != 3 {
		t.Errorf("expected 3 channel summaries, got %d", len(summary.Channels))
	}
}

func TestWriteSummaryReport_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	err := WriteSummaryReport(tmpDir, nil)
	if err != nil {
		t.Fatalf("write summary report: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "migration-report.yaml"))
	if err != nil {
		t.Fatalf("read summary report: %v", err)
	}

	var summary SummaryReport
	if err := yaml.Unmarshal(data, &summary); err != nil {
		t.Fatalf("unmarshal summary report: %v", err)
	}

	if summary.TotalChannels != 0 {
		t.Errorf("expected total 0, got %d", summary.TotalChannels)
	}
}
