package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sroopra/ghega/pkg/mapping"
	"gopkg.in/yaml.v3"
)

// MigrationStatus represents the overall status of a channel migration.
type MigrationStatus string

const (
	StatusAutoConverted MigrationStatus = "auto-converted"
	StatusNeedsRewrite  MigrationStatus = "needs-rewrite"
	StatusUnsupported   MigrationStatus = "unsupported"
	StatusMixed         MigrationStatus = "mixed"
)

// TypedRewriteTask extends RewriteTask with classification context.
type TypedRewriteTask struct {
	StepName    string          `yaml:"stepName"`
	Category    PatternCategory `yaml:"category"`
	Severity    string          `yaml:"severity"`
	Description string          `yaml:"description"`
}

// ChannelMigrationReport is the per-channel report.
type ChannelMigrationReport struct {
	ChannelName   string          `yaml:"channelName"`
	Status        MigrationStatus `yaml:"status"`
	AutoConverted []string        `yaml:"autoConverted,omitempty"`
	NeedsRewrite  []RewriteTask   `yaml:"needsRewrite,omitempty"`
	Unsupported   []string        `yaml:"unsupported,omitempty"`
	Warnings      []string        `yaml:"warnings,omitempty"`
}

// ChannelSummary is a brief entry for each channel in the summary.
type ChannelSummary struct {
	Name              string          `yaml:"name"`
	Status            MigrationStatus `yaml:"status"`
	WarningsCount     int             `yaml:"warningsCount"`
	RewriteTasksCount int             `yaml:"rewriteTasksCount"`
}

// SummaryReport is the root migration report.
type SummaryReport struct {
	GeneratedAt   time.Time        `yaml:"generatedAt"`
	TotalChannels int              `yaml:"totalChannels"`
	AutoConverted int              `yaml:"autoConverted"`
	NeedsRewrite  int              `yaml:"needsRewrite"`
	Unsupported   int              `yaml:"unsupported"`
	Mixed         int              `yaml:"mixed"`
	Channels      []ChannelSummary `yaml:"channels"`
}

// GenerateMigrationReport processes a single converted channel and its
// classified transformer steps and produces the per-channel artifacts.
func GenerateMigrationReport(
	channelName string,
	convResult *ConversionResult,
	classifications []ClassificationResult,
) (*ChannelMigrationReport, []TypedRewriteTask, []mapping.Mapping) {
	report := &ChannelMigrationReport{
		ChannelName:   channelName,
		AutoConverted: make([]string, 0),
		NeedsRewrite:  make([]RewriteTask, 0),
		Unsupported:   make([]string, 0),
		Warnings:      append([]string(nil), convResult.Warnings...),
	}

	var rewriteTasks []TypedRewriteTask
	var autoMappings []mapping.Mapping

	// Record connector conversions.
	if convResult.Channel.Source.Type != "" {
		report.AutoConverted = append(report.AutoConverted,
			fmt.Sprintf("Source connector converted to %s", convResult.Channel.Source.Type))
	}
	if convResult.Channel.Destination.Type != "" {
		report.AutoConverted = append(report.AutoConverted,
			fmt.Sprintf("Destination connector converted to %s", convResult.Channel.Destination.Type))
	}

	for _, cr := range classifications {
		for _, p := range cr.Patterns {
			switch p.Disposition {
			case DispositionAutoConvertible:
				report.AutoConverted = append(report.AutoConverted, p.Description)
				if p.Mapping != nil {
					autoMappings = append(autoMappings, *p.Mapping)
				}
			case DispositionNeedsRewrite:
				report.NeedsRewrite = append(report.NeedsRewrite, *p.RewriteTask)
				rewriteTasks = append(rewriteTasks, TypedRewriteTask{
					StepName:    cr.Step.Name,
					Category:    p.Category,
					Severity:    p.RewriteTask.Severity,
					Description: p.RewriteTask.Description,
				})
			case DispositionUnsupported:
				report.Unsupported = append(report.Unsupported, p.Description)
				if p.RewriteTask != nil {
					rewriteTasks = append(rewriteTasks, TypedRewriteTask{
						StepName:    cr.Step.Name,
						Category:    p.Category,
						Severity:    p.RewriteTask.Severity,
						Description: p.RewriteTask.Description,
					})
				}
			}
		}
	}

	// Promote unsupported-connector warnings to unsupported entries.
	for _, w := range report.Warnings {
		if strings.Contains(w, "unsupported source connector type") || strings.Contains(w, "unsupported destination connector type") {
			report.Unsupported = append(report.Unsupported, w)
		}
	}

	report.Status = deriveMigrationStatus(report)
	return report, rewriteTasks, autoMappings
}

func deriveMigrationStatus(report *ChannelMigrationReport) MigrationStatus {
	hasRewrite := len(report.NeedsRewrite) > 0
	hasUnsupported := len(report.Unsupported) > 0

	if hasUnsupported {
		return StatusUnsupported
	}
	if hasRewrite {
		return StatusNeedsRewrite
	}
	return StatusAutoConverted
}

// WriteChannelOutput writes the per-channel artifacts to disk.
func WriteChannelOutput(
	outDir string,
	channelName string,
	convResult *ConversionResult,
	report *ChannelMigrationReport,
	rewriteTasks []TypedRewriteTask,
	autoMappings []mapping.Mapping,
) error {
	dir := filepath.Join(outDir, channelName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create channel dir: %w", err)
	}

	// Write channel.yaml with auto-mappings merged.
	ch := convResult.Channel
	ch.Mappings = append(ch.Mappings, autoMappings...)

	chData, err := yaml.Marshal(&ch)
	if err != nil {
		return fmt.Errorf("marshal channel: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "channel.yaml"), chData, 0644); err != nil {
		return fmt.Errorf("write channel.yaml: %w", err)
	}

	// Write rewrite-tasks.yaml.
	rtData, err := yaml.Marshal(rewriteTasks)
	if err != nil {
		return fmt.Errorf("marshal rewrite tasks: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rewrite-tasks.yaml"), rtData, 0644); err != nil {
		return fmt.Errorf("write rewrite-tasks.yaml: %w", err)
	}

	// Write migration-report.yaml.
	rptData, err := yaml.Marshal(report)
	if err != nil {
		return fmt.Errorf("marshal migration report: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "migration-report.yaml"), rptData, 0644); err != nil {
		return fmt.Errorf("write migration-report.yaml: %w", err)
	}

	return nil
}

// WriteSummaryReport writes the root summary report.
func WriteSummaryReport(outDir string, reports []*ChannelMigrationReport) error {
	summary := SummaryReport{
		GeneratedAt:   time.Now().UTC(),
		TotalChannels: len(reports),
		Channels:      make([]ChannelSummary, 0, len(reports)),
	}

	for _, r := range reports {
		summary.Channels = append(summary.Channels, ChannelSummary{
			Name:              r.ChannelName,
			Status:            r.Status,
			WarningsCount:     len(r.Warnings),
			RewriteTasksCount: len(r.NeedsRewrite),
		})
		switch r.Status {
		case StatusAutoConverted:
			summary.AutoConverted++
		case StatusNeedsRewrite:
			summary.NeedsRewrite++
		case StatusUnsupported:
			summary.Unsupported++
		case StatusMixed:
			summary.Mixed++
		}
	}

	data, err := yaml.Marshal(&summary)
	if err != nil {
		return fmt.Errorf("marshal summary: %w", err)
	}
	return os.WriteFile(filepath.Join(outDir, "migration-report.yaml"), data, 0644)
}
