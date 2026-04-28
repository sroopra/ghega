package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sroopra/ghega/pkg/channel"
	"github.com/sroopra/ghega/pkg/channelstore"
	"gopkg.in/yaml.v3"
)

// channelStoreOverride is set by tests to inject a shared store.
var channelStoreOverride channelstore.ChannelStore

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

func runChannelDeploy(args []string) error {
	fs := flag.NewFlagSet("deploy", flag.ExitOnError)
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega channel deploy <path-to-channel.yaml>")
	}

	channelPath := fs.Arg(0)
	store, err := initChannelStore()
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	res, err := channel.Deploy(channelPath, store)
	if err != nil {
		return fmt.Errorf("deploy: %w", err)
	}

	if res.IsNoOp {
		fmt.Printf("%s is already at revision %d hash %s\n", res.Name, res.Revision, res.Hash)
	} else {
		fmt.Printf("Deployed %s revision %d hash %s\n", res.Name, res.Revision, res.Hash)
	}
	return nil
}

func runChannelDiff(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega channel diff <path-to-channel.yaml>")
	}

	channelPath := fs.Arg(0)
	store, err := initChannelStore()
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	res, err := channel.DiffLocal(channelPath, store)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}

	if res.DeployedHash == "" {
		fmt.Printf("Channel %s has never been deployed\n", res.ChannelName)
		return nil
	}

	if res.Identical {
		fmt.Printf("No changes — local matches deployed revision %s\n", res.DeployedHash)
	} else {
		fmt.Printf("Local:  %s\n", res.LocalHash)
		fmt.Printf("Deployed: %s\n", res.DeployedHash)
		fmt.Printf("Channel %s has un-deployed changes\n", res.ChannelName)
	}
	return nil
}

func runChannelRollback(args []string) error {
	fs := flag.NewFlagSet("rollback", flag.ExitOnError)
	toHash := fs.String("to", "", "hash to roll back to")
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega channel rollback <name> --to <hash>")
	}

	name := fs.Arg(0)
	store, err := initChannelStore()
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	if err := channel.Rollback(name, *toHash, store); err != nil {
		return fmt.Errorf("rollback: %w", err)
	}

	fmt.Printf("Rolled back %s to hash %s\n", name, *toHash)
	return nil
}

func initChannelStore() (channelstore.ChannelStore, error) {
	if channelStoreOverride != nil {
		return channelStoreOverride, nil
	}

	dsn := os.Getenv("GHEGA_DATABASE_URL")
	if dsn == "" {
		dsn = "ghega.db"
	}

	if dsn == ":memory:" {
		return channelstore.NewInMemoryStore(), nil
	}

	store, err := channelstore.NewSQLiteStore(dsn)
	if err != nil {
		return channelstore.NewInMemoryStore(), nil
	}
	return store, nil
}
