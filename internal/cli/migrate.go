package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sroopra/ghega/pkg/migration"
	"github.com/sroopra/ghega/pkg/mirthxml"
)

func runMigrate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ghega migrate <subcommand>")
	}

	switch args[0] {
	case "mirth":
		return runMigrateMirth(args[1:])
	default:
		return fmt.Errorf("unknown migrate subcommand: %s", args[0])
	}
}

func parseMigrateMirthArgs(args []string) (exportDir, outDir, samplesDir, expectedDir string, err error) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--out":
			if i+1 >= len(args) {
				return "", "", "", "", fmt.Errorf("missing value for --out")
			}
			outDir = args[i+1]
			i++
		case "--samples":
			if i+1 >= len(args) {
				return "", "", "", "", fmt.Errorf("missing value for --samples")
			}
			samplesDir = args[i+1]
			i++
		case "--expected":
			if i+1 >= len(args) {
				return "", "", "", "", fmt.Errorf("missing value for --expected")
			}
			expectedDir = args[i+1]
			i++
		default:
			if strings.HasPrefix(args[i], "-") {
				return "", "", "", "", fmt.Errorf("unknown flag: %s", args[i])
			}
			if exportDir == "" {
				exportDir = args[i]
			} else {
				return "", "", "", "", fmt.Errorf("unexpected argument: %s", args[i])
			}
		}
	}
	if exportDir == "" {
		return "", "", "", "", fmt.Errorf("usage: ghega migrate mirth <export-dir> --out <output-dir>")
	}
	if outDir == "" {
		return "", "", "", "", fmt.Errorf("usage: ghega migrate mirth <export-dir> --out <output-dir>")
	}
	return exportDir, outDir, samplesDir, expectedDir, nil
}

func runMigrateMirth(args []string) error {
	exportDir, outDir, samplesDir, expectedDir, err := parseMigrateMirthArgs(args)
	if err != nil {
		return err
	}

	// Create output directory.
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// Walk export directory for .xml files.
	entries, err := os.ReadDir(exportDir)
	if err != nil {
		return fmt.Errorf("read export directory: %w", err)
	}

	var allReports []*migration.ChannelMigrationReport
	seenNames := make(map[string]int)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".xml" {
			continue
		}

		path := filepath.Join(exportDir, entry.Name())
		ch, err := mirthxml.ParseChannelFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skipped %s: %v\n", entry.Name(), err)
			continue
		}

		// Convert channel.
		convResult, err := migration.ConvertChannel(ch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to convert %s: %v\n", entry.Name(), err)
			continue
		}

		// Disambiguate duplicate sanitized names.
		channelName := convResult.Channel.Name
		if seenNames[channelName] > 0 {
			channelName = fmt.Sprintf("%s-%d", channelName, seenNames[channelName])
		}
		seenNames[convResult.Channel.Name]++

		// Add warnings for filter rules (not converted in this step).
		if len(ch.SourceConnector.Filter.Rules) > 0 {
			convResult.Warnings = append(convResult.Warnings,
				fmt.Sprintf("source filter has %d rule(s) that require manual review", len(ch.SourceConnector.Filter.Rules)))
		}
		for _, dest := range ch.DestinationConnectors {
			if len(dest.Filter.Rules) > 0 {
				convResult.Warnings = append(convResult.Warnings,
					fmt.Sprintf("destination %q filter has %d rule(s) that require manual review", dest.Name, len(dest.Filter.Rules)))
			}
		}

		// Classify all transformer steps.
		var classifications []migration.ClassificationResult
		for _, step := range ch.SourceConnector.Transformer.Steps {
			classifications = append(classifications, migration.ClassifyTransformerStep(step))
		}
		for _, dest := range ch.DestinationConnectors {
			for _, step := range dest.Transformer.Steps {
				classifications = append(classifications, migration.ClassifyTransformerStep(step))
			}
		}

		// Generate report.
		report, rewriteTasks, autoMappings := migration.GenerateMigrationReport(
			channelName,
			convResult,
			classifications,
		)

		// Write per-channel output.
		if err := migration.WriteChannelOutput(
			outDir,
			channelName,
			convResult,
			report,
			rewriteTasks,
			autoMappings,
		); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write output for %s: %v\n", entry.Name(), err)
			continue
		}

		allReports = append(allReports, report)
	}

	// Write summary report.
	if err := migration.WriteSummaryReport(outDir, allReports); err != nil {
		return fmt.Errorf("write summary report: %w", err)
	}

	// Stubs for optional flags.
	if samplesDir != "" {
		fmt.Fprintf(os.Stderr, "samples flag is not yet implemented\n")
	}
	if expectedDir != "" {
		fmt.Fprintf(os.Stderr, "expected flag is not yet implemented\n")
	}

	fmt.Printf("Migration complete. Output written to %s\n", outDir)
	return nil
}
