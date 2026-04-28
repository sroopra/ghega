package cli

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sroopra/ghega/pkg/channel"
)

func runWatch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ghega watch <directory>")
	}

	dir := args[0]
	mtimes := make(map[string]time.Time)

	fmt.Printf("Watching %s for channel.yaml changes...\n", dir)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	for {
		select {
		case <-ticker.C:
			if err := scanAndProcess(dir, mtimes); err != nil {
				fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
			}
		case <-sigCh:
			fmt.Println("\nReceived interrupt signal. Stopping watch.")
			return nil
		}
	}
}

func scanAndProcess(dir string, mtimes map[string]time.Time) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || info.Name() != "channel.yaml" {
			return nil
		}

		prev, seen := mtimes[path]
		mtimes[path] = info.ModTime()
		if !seen || !info.ModTime().Equal(prev) {
			fmt.Printf("\n[watch] Change detected: %s\n", path)
			validateAndTest(path)
		}
		return nil
	})
}

func validateAndTest(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("\033[31mERROR  %s: %v\033[0m\n", path, err)
		return
	}

	ch, valErrs := channel.ValidateYAML(data)
	if ch != nil {
		valErrs = append(valErrs, channel.ValidatePolicies(ch)...)
	}
	if len(valErrs) > 0 {
		fmt.Printf("\033[31mVALIDATE FAIL %s\033[0m\n", path)
		for _, e := range valErrs {
			fmt.Printf("  %s: %s\n", e.Field, e.Message)
		}
		return
	}

	fmt.Printf("\033[32mVALIDATE PASS %s\033[0m\n", path)

	fixtures, err := channel.LoadTestFixtures(path, ch.Tests)
	if err != nil {
		fmt.Printf("\033[31mTEST FAIL %s: load fixtures: %v\033[0m\n", path, err)
		return
	}

	allPassed := true
	for _, fixture := range fixtures {
		result, err := channel.RunTest(fixture, ch.Mappings)
		if err != nil {
			fmt.Printf("\033[31mTEST FAIL %s/%s: %v\033[0m\n", path, fixture.Name, err)
			allPassed = false
			continue
		}
		if result.Passed {
			fmt.Printf("\033[32mTEST PASS %s/%s\033[0m\n", path, result.Name)
		} else {
			fmt.Printf("\033[31mTEST FAIL %s/%s: %s\033[0m\n", path, result.Name, strings.Join(result.Errors, "; "))
			allPassed = false
		}
	}

	if allPassed {
		fmt.Printf("\033[32mALL PASS %s\033[0m\n", path)
	}
}
