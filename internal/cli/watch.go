package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sroopra/ghega/pkg/channel"
	"gopkg.in/yaml.v3"
)

const (
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

// runWatch implements ghega watch <directory>.
func runWatch(args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("stat directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	state := make(map[string]time.Time)

	// Initial scan.
	if err := scanAndRun(dir, state); err != nil {
		return err
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Printf("Watching %s for channel.yaml changes (press Ctrl+C to stop)\n", dir)

	for {
		select {
		case <-ticker.C:
			if err := scanAndRun(dir, state); err != nil {
				fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
			}
		case <-ctx.Done():
			fmt.Println("\nStopping watch.")
			return nil
		}
	}
}

func scanAndRun(dir string, state map[string]time.Time) error {
	files, err := findChannelYAMLs(dir)
	if err != nil {
		return err
	}

	for _, path := range files {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		mtime := info.ModTime()
		last, ok := state[path]
		if !ok {
			state[path] = mtime
			continue
		}

		if mtime.After(last) {
			state[path] = mtime
			fmt.Printf("\n[change detected] %s\n", path)
			if err := validateAndTest(path); err != nil {
				fmt.Fprintf(os.Stderr, "  %sFAIL%s %v\n", colorRed, colorReset, err)
			} else {
				fmt.Printf("  %sPASS%s validate + test\n", colorGreen, colorReset)
			}
		}
	}

	return nil
}

func findChannelYAMLs(dir string) ([]string, error) {
	var out []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable paths
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == "channel.yaml" {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

func validateAndTest(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	ch, valErrs := channel.ValidateYAML(data)
	if ch != nil {
		valErrs = append(valErrs, channel.ValidatePolicies(ch)...)
	}
	if len(valErrs) > 0 {
		var msgs []string
		for _, e := range valErrs {
			msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
		}
		return fmt.Errorf("validation failed: %s", strings.Join(msgs, "; "))
	}

	var chParsed channel.Channel
	if err := yaml.Unmarshal(data, &chParsed); err != nil {
		return fmt.Errorf("parse channel yaml: %w", err)
	}

	fixtures, err := channel.LoadTestFixtures(path, chParsed.Tests)
	if err != nil {
		return fmt.Errorf("load test fixtures: %w", err)
	}

	for _, fixture := range fixtures {
		result, err := channel.RunTest(fixture, chParsed.Mappings)
		if err != nil {
			return fmt.Errorf("run test %q: %w", fixture.Name, err)
		}
		if result.Passed {
			fmt.Printf("  %sPASS%s %s\n", colorGreen, colorReset, result.Name)
		} else {
			fmt.Printf("  %sFAIL%s %s: %s\n", colorRed, colorReset, result.Name, strings.Join(result.Errors, "; "))
		}
		for _, w := range result.Warnings {
			fmt.Printf("  WARN %s: %s\n", result.Name, w)
		}
	}

	return nil
}
