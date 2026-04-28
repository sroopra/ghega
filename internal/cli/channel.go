package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sroopra/ghega/pkg/channel"
	"github.com/sroopra/ghega/pkg/channelstore"
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
	case "deploy":
		return runChannelDeploy(args[1:])
	case "diff":
		return runChannelDiff(args[1:])
	case "rollback":
		return runChannelRollback(args[1:])
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
		return fmt.Errorf("error reading file: %w", err)
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
		return fmt.Errorf("validation failed:\n%s", strings.Join(msgs, "\n"))
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

	ch, valErrs := channel.ValidateYAML(data)
	if ch != nil {
		valErrs = append(valErrs, channel.ValidatePolicies(ch)...)
	}
	if len(valErrs) > 0 {
		var msgs []string
		for _, e := range valErrs {
			msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
		}
		return fmt.Errorf("validation failed:\n%s", strings.Join(msgs, "\n"))
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
		return fmt.Errorf("one or more tests failed")
	}

	return nil
}

func initChannelStore() (channelstore.ChannelStore, error) {
	dsn := os.Getenv("GHEGA_DATABASE_URL")
	if dsn == "" {
		dsn = "ghega.db"
	}
	return channelstore.NewSQLiteStore(dsn)
}

func runChannelDeploy(args []string) error {
	fs := flag.NewFlagSet("deploy", flag.ExitOnError)
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega channel deploy <path-to-channel.yaml>")
	}

	store, err := initChannelStore()
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	result, err := channel.Deploy(fs.Arg(0), store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "deploy failed: %v\n", err)
		os.Exit(1)
	}

	if result.PreviousHash == "" {
		fmt.Printf("Deployed %s revision %d hash %s\n", result.Name, result.Revision, result.Hash)
	} else if result.PreviousHash == result.Hash {
		fmt.Printf("%s is already at revision %d hash %s\n", result.Name, result.Revision, result.Hash)
	} else {
		fmt.Printf("Deployed %s revision %d hash %s (previous: %s)\n", result.Name, result.Revision, result.Hash, result.PreviousHash)
	}
	return nil
}

func runChannelDiff(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega channel diff <path-to-channel.yaml>")
	}

	store, err := initChannelStore()
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	result, err := channel.DiffLocal(fs.Arg(0), store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "diff failed: %v\n", err)
		os.Exit(1)
	}

	if result.DeployedHash == "" {
		fmt.Printf("Channel %s has never been deployed\n", result.ChannelName)
	} else if result.Identical {
		fmt.Printf("No changes — local matches deployed revision %s\n", result.DeployedHash)
	} else {
		fmt.Printf("Local:    %s\nDeployed: %s\nChannel %s has un-deployed changes.\n", result.LocalHash, result.DeployedHash, result.ChannelName)
	}
	return nil
}

func runChannelRollback(args []string) error {
	fs := flag.NewFlagSet("rollback", flag.ExitOnError)
	toHash := fs.String("to", "", "hash to rollback to")
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega channel rollback <name> --to <hash>")
	}

	store, err := initChannelStore()
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	name := fs.Arg(0)
	if err := channel.Rollback(name, *toHash, store); err != nil {
		fmt.Fprintf(os.Stderr, "rollback failed: %v\n", err)
		os.Exit(1)
	}

	hash := *toHash
	if hash == "" {
		rec, err := store.GetChannel(context.Background(), name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rollback failed: %v\n", err)
			os.Exit(1)
		}
		hash = rec.Hash
	}

	fmt.Printf("Rolled back %s to hash %s\n", name, hash)
	return nil
}
