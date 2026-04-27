package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sroopra/ghega/pkg/channel"
	"gopkg.in/yaml.v3"
)

func runChannel(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ghega channel <subcommand> [flags]")
	}

	switch args[0] {
	case "validate":
		return runChannelValidate(args[1:])
	case "test":
		return runChannelTest(args[1:])
	default:
		return fmt.Errorf("unknown channel subcommand: %s", args[0])
	}
}

func runChannelValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega channel validate <path>")
	}

	path := fs.Arg(0)
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}

	ch, valErrs := channel.ValidateYAML(data)
	if ch != nil {
		valErrs = append(valErrs, channel.ValidatePolicies(ch)...)
	}

	if len(valErrs) > 0 {
		for _, e := range valErrs {
			fmt.Fprintf(os.Stderr, "%s: %s\n", e.Field, e.Message)
		}
		os.Exit(1)
	}

	fmt.Println("channel is valid")
	return nil
}

func runChannelTest(args []string) error {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	junitPath := fs.String("junit", "", "write JUnit XML report to path")
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega channel test <path-to-channel.yaml>")
	}

	channelPath := fs.Arg(0)
	data, err := os.ReadFile(channelPath)
	if err != nil {
		return fmt.Errorf("read channel file: %w", err)
	}

	var ch channel.Channel
	if err := yaml.Unmarshal(data, &ch); err != nil {
		return fmt.Errorf("parse channel yaml: %w", err)
	}

	fixtures, err := channel.LoadTestFixtures(channelPath, ch.Tests)
	if err != nil {
		return fmt.Errorf("load test fixtures: %w", err)
	}

	var results []*channel.TestResult
	allPassed := true

	for _, fixture := range fixtures {
		result, err := channel.RunTest(fixture, ch.Mappings)
		if err != nil {
			return fmt.Errorf("run test %q: %w", fixture.Name, err)
		}
		results = append(results, result)
		if !result.Passed {
			allPassed = false
		}
	}

	// Print results.
	for _, r := range results {
		if r.Passed {
			fmt.Printf("PASS %s\n", r.Name)
		} else {
			fmt.Printf("FAIL %s: %s\n", r.Name, strings.Join(r.Errors, "; "))
		}
		for _, w := range r.Warnings {
			fmt.Printf("WARN %s: %s\n", r.Name, w)
		}
	}

	// Write JUnit report if requested.
	if *junitPath != "" {
		typed := make([]channel.TestResult, len(results))
		for i, r := range results {
			typed[i] = *r
		}
		report := channel.ToJUnit(typed)
		if err := os.WriteFile(*junitPath, []byte(report), 0644); err != nil {
			return fmt.Errorf("write junit report: %w", err)
		}
	}

	if !allPassed {
		os.Exit(1)
	}

	return nil
}
