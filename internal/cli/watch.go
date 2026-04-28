package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sroopra/ghega/pkg/channel"
)

const (
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

// watchNoColor is set by tests to disable ANSI escape codes.
var watchNoColor = false

func runWatch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ghega watch <directory>")
	}

	dir := args[0]
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	mtimes := make(map[string]time.Time)

	// Initial scan.
	if err := scanAndUpdate(dir, mtimes); err != nil {
		return err
	}

	fmt.Printf("Watching %s for channel.yaml changes...\n", dir)

	for {
		select {
		case <-ticker.C:
			changed, err := findChanged(dir, mtimes)
			if err != nil {
				fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
				continue
			}
			for _, path := range changed {
				runWatchValidation(path)
			}
		}
	}
}

func scanAndUpdate(dir string, mtimes map[string]time.Time) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "channel.yaml") {
			mtimes[path] = info.ModTime()
		}
		return nil
	})
}

func findChanged(dir string, mtimes map[string]time.Time) ([]string, error) {
	var changed []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, "channel.yaml") {
			return nil
		}

		old, ok := mtimes[path]
		if !ok || !info.ModTime().Equal(old) {
			mtimes[path] = info.ModTime()
			changed = append(changed, path)
		}
		return nil
	})

	return changed, err
}

func runWatchValidation(path string) {
	green, red, reset := colorGreen, colorRed, colorReset
	if watchNoColor {
		green, red, reset = "", "", ""
	}

	fmt.Printf("\n==> %s\n", path)

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("%sFAIL%s read: %v\n", red, reset, err)
		return
	}

	ch, valErrs := channel.ValidateYAML(data)
	if ch != nil {
		valErrs = append(valErrs, channel.ValidatePolicies(ch)...)
	}

	if len(valErrs) > 0 {
		fmt.Printf("%sFAIL%s validation:\n", red, reset)
		for _, e := range valErrs {
			fmt.Printf("  %s: %s\n", e.Field, e.Message)
		}
		return
	}

	fmt.Printf("%sPASS%s validation\n", green, reset)

	fixtures, err := channel.LoadTestFixtures(path, ch.Tests)
	if err != nil {
		fmt.Printf("%sFAIL%s fixtures: %v\n", red, reset, err)
		return
	}

	allPassed := true
	for _, fixture := range fixtures {
		result, err := channel.RunTest(fixture, ch.Mappings)
		if err != nil {
			fmt.Printf("%sFAIL%s %s: %v\n", red, reset, fixture.Name, err)
			allPassed = false
			continue
		}
		if result.Passed {
			fmt.Printf("%sPASS%s %s\n", green, reset, fixture.Name)
		} else {
			fmt.Printf("%sFAIL%s %s: %s\n", red, reset, fixture.Name, strings.Join(result.Errors, "; "))
			allPassed = false
		}
	}

	if allPassed {
		fmt.Printf("%sPASS%s all tests\n", green, reset)
	}
}
