package cli

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/sroopra/ghega/pkg/migration"
)

func runMigrate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ghega migrate <subcommand> [flags]")
	}

	switch args[0] {
	case "mirth":
		return runMigrateMirth(args[1:])
	default:
		return fmt.Errorf("unknown migrate subcommand: %s", args[0])
	}
}

func runMigrateMirth(args []string) error {
	fs := flag.NewFlagSet("mirth", flag.ExitOnError)
	out := fs.String("out", "", "Output directory for migration reports (required)")
	samples := fs.String("samples", "", "Directory containing sample messages for validation (optional)")
	expected := fs.String("expected", "", "Directory containing expected golden files for comparison (optional)")

	// Re-order args so that flags may appear before or after positional args.
	if err := fs.Parse(reorderFlags(args)); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega migrate mirth <export-dir> --out <output-dir>")
	}

	exportDir := fs.Arg(0)
	if *out == "" {
		return fmt.Errorf("--out is required")
	}

	// --samples and --expected are reserved for future golden-file testing.
	if *samples != "" || *expected != "" {
		slog.Warn("--samples and --expected are reserved for future use and currently have no effect")
	}

	_, err := migration.GenerateMigrationReports(exportDir, *out)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Printf("Migration reports written to %s\n", *out)
	return nil
}

// reorderFlags moves flag arguments (and their values) to the front of the
// slice so that flag.FlagSet.Parse can locate them even when positional
// arguments are given first.
func reorderFlags(args []string) []string {
	known := map[string]bool{
		"--out": true, "-out": true,
		"--samples": true, "-samples": true,
		"--expected": true, "-expected": true,
	}
	var flags []string
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if known[arg] && i+1 < len(args) {
			flags = append(flags, arg, args[i+1])
			i++
			continue
		}
		positional = append(positional, arg)
	}
	return append(flags, positional...)
}
