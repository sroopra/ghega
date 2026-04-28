// Package migration generates migration reports from Mirth channels to Ghega.
package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/sroopra/ghega/pkg/channel"
	"github.com/sroopra/ghega/pkg/mapping"
	"github.com/sroopra/ghega/pkg/mirthxml"
)

// ChannelMigrationReport is the per-channel migration report.
type ChannelMigrationReport struct {
	ChannelName   string            `yaml:"channelName"`
	OriginalName  string            `yaml:"originalName"`
	Status        string            `yaml:"status"`
	AutoConverted []ConvertedItem   `yaml:"autoConverted"`
	NeedsRewrite  []RewriteTaskItem `yaml:"needsRewrite"`
	Unsupported   []UnsupportedItem `yaml:"unsupported"`
	Warnings      []string          `yaml:"warnings"`
}

// ConvertedItem describes a successfully migrated element.
type ConvertedItem struct {
	Element     string `yaml:"element"`
	Description string `yaml:"description"`
}

// RewriteTaskItem describes a typed rewrite task.
type RewriteTaskItem struct {
	Severity    string `yaml:"severity"`
	Description string `yaml:"description"`
	Category    string `yaml:"category,omitempty"`
}

// UnsupportedItem describes a feature that is not yet supported.
type UnsupportedItem struct {
	Feature     string `yaml:"feature"`
	Description string `yaml:"description"`
}

// SummaryReport is the top-level migration summary.
type SummaryReport struct {
	GeneratedAt        string           `yaml:"generatedAt"`
	Channels           []ChannelSummary `yaml:"channels"`
	TotalChannels      int              `yaml:"totalChannels"`
	TotalAutoConverted int              `yaml:"totalAutoConverted"`
	TotalNeedsRewrite  int              `yaml:"totalNeedsRewrite"`
	TotalUnsupported   int              `yaml:"totalUnsupported"`
}

// ChannelSummary is a condensed view of a single channel's migration result.
type ChannelSummary struct {
	Name              string `yaml:"name"`
	Status            string `yaml:"status"`
	RewriteTasksCount int    `yaml:"rewriteTasksCount"`
	WarningsCount     int    `yaml:"warningsCount"`
}

// GenerateMigrationReports walks exportDir for Mirth channel XML files,
// converts each to a Ghega channel, classifies transformers, and writes
// per-channel reports plus a summary report to outDir.
func GenerateMigrationReports(exportDir, outDir string) (*SummaryReport, error) {
	channels, err := mirthxml.ParseChannelsFromDir(exportDir)
	if err != nil {
		return nil, fmt.Errorf("parse channels from dir: %w", err)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	summary := &SummaryReport{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Channels:    []ChannelSummary{},
	}

	for _, mch := range channels {
		report, err := processChannel(mch, outDir)
		if err != nil {
			return nil, fmt.Errorf("process channel %s: %w", mch.Name, err)
		}

		summary.Channels = append(summary.Channels, ChannelSummary{
			Name:              report.ChannelName,
			Status:            report.Status,
			RewriteTasksCount: len(report.NeedsRewrite),
			WarningsCount:     len(report.Warnings),
		})
		summary.TotalChannels++
		switch report.Status {
		case "auto-converted":
			summary.TotalAutoConverted++
		case "needs-rewrite", "mixed":
			summary.TotalNeedsRewrite++
		case "unsupported":
			summary.TotalUnsupported++
		}
	}

	summaryPath := filepath.Join(outDir, "migration-report.yaml")
	data, err := yaml.Marshal(summary)
	if err != nil {
		return nil, fmt.Errorf("marshal summary report: %w", err)
	}
	if err := os.WriteFile(summaryPath, data, 0644); err != nil {
		return nil, fmt.Errorf("write summary report: %w", err)
	}

	return summary, nil
}

func processChannel(mch *mirthxml.Channel, outDir string) (*ChannelMigrationReport, error) {
	convResult, err := ConvertChannel(mch)
	if err != nil {
		return nil, fmt.Errorf("convert channel: %w", err)
	}

	report := &ChannelMigrationReport{
		ChannelName:  convResult.Channel.Name,
		OriginalName: mch.Name,
		Warnings:     append([]string{}, convResult.Warnings...),
	}

	// Record auto-converted structural elements.
	if convResult.Channel.Source.Type != "" {
		report.AutoConverted = append(report.AutoConverted, ConvertedItem{
			Element:     "source_connector",
			Description: fmt.Sprintf("Source connector mapped to type %q", convResult.Channel.Source.Type),
		})
	}
	if convResult.Channel.Destination.Type != "" {
		report.AutoConverted = append(report.AutoConverted, ConvertedItem{
			Element:     "destination_connector",
			Description: fmt.Sprintf("Destination connector mapped to type %q", convResult.Channel.Destination.Type),
		})
	}

	var mappings []mapping.Mapping

	// Classify source transformer steps.
	for _, step := range mch.SourceConnector.Transformer.Steps {
		cr := ClassifyTransformerStep(step)
		mergeClassification(report, cr, &mappings)
	}

	// Classify destination transformer steps.
	for _, dest := range mch.DestinationConnectors {
		for _, step := range dest.Transformer.Steps {
			cr := ClassifyTransformerStep(step)
			mergeClassification(report, cr, &mappings)
		}
	}

	// Record auto-converted mappings.
	for _, m := range mappings {
		report.AutoConverted = append(report.AutoConverted, ConvertedItem{
			Element:     "mapping",
			Description: fmt.Sprintf("Mapping from %s to %s (%s)", m.Source, m.Target, m.Transform),
		})
	}

	// Derive overall status.
	if len(report.Unsupported) > 0 && len(report.NeedsRewrite) > 0 {
		report.Status = "mixed"
	} else if len(report.Unsupported) > 0 {
		report.Status = "unsupported"
	} else if len(report.NeedsRewrite) > 0 {
		report.Status = "needs-rewrite"
	} else {
		report.Status = "auto-converted"
	}

	channelDir := filepath.Join(outDir, report.ChannelName)
	if err := os.MkdirAll(channelDir, 0755); err != nil {
		return nil, fmt.Errorf("create channel directory: %w", err)
	}

	// Write channel.yaml
	convResult.Channel.Mappings = mappings
	chData, err := yaml.Marshal(convResult.Channel)
	if err != nil {
		return nil, fmt.Errorf("marshal channel: %w", err)
	}
	if err := os.WriteFile(filepath.Join(channelDir, "channel.yaml"), chData, 0644); err != nil {
		return nil, fmt.Errorf("write channel.yaml: %w", err)
	}

	// Write rewrite-tasks.yaml (always write, even if empty).
	rtData, err := yaml.Marshal(report.NeedsRewrite)
	if err != nil {
		return nil, fmt.Errorf("marshal rewrite tasks: %w", err)
	}
	if err := os.WriteFile(filepath.Join(channelDir, "rewrite-tasks.yaml"), rtData, 0644); err != nil {
		return nil, fmt.Errorf("write rewrite-tasks.yaml: %w", err)
	}

	// Write per-channel migration-report.yaml
	rptData, err := yaml.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("marshal report: %w", err)
	}
	if err := os.WriteFile(filepath.Join(channelDir, "migration-report.yaml"), rptData, 0644); err != nil {
		return nil, fmt.Errorf("write migration-report.yaml: %w", err)
	}

	return report, nil
}

func mergeClassification(report *ChannelMigrationReport, cr ClassificationResult, mappings *[]mapping.Mapping) {
	for _, p := range cr.Patterns {
		switch p.Disposition {
		case DispositionAutoConvertible:
			if p.Mapping != nil {
				*mappings = append(*mappings, *p.Mapping)
			}
		case DispositionNeedsRewrite:
			task := RewriteTaskItem{
				Severity:    "medium",
				Description: p.Description,
				Category:    string(p.Category),
			}
			if p.RewriteTask != nil {
				task.Severity = p.RewriteTask.Severity
				task.Description = p.RewriteTask.Description
			}
			report.NeedsRewrite = append(report.NeedsRewrite, task)
		case DispositionUnsupported:
			item := UnsupportedItem{
				Feature:     string(p.Category),
				Description: p.Description,
			}
			report.Unsupported = append(report.Unsupported, item)
		}
	}
}

// WriteChannelYAML serialises a Ghega channel to a YAML file.
func WriteChannelYAML(ch channel.Channel, path string) error {
	data, err := yaml.Marshal(ch)
	if err != nil {
		return fmt.Errorf("marshal channel: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write channel file: %w", err)
	}
	return nil
}
