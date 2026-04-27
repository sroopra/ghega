package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/sroopra/ghega/pkg/channel"
	"github.com/sroopra/ghega/pkg/channelstore"
)

func runChannel(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf(`usage: ghega channel <subcommand> [args]

Subcommands:
  validate   Validate a channel definition
  deploy     Deploy a channel
  diff       Compare local channel to deployed version
  rollback   Roll back a channel to a previous revision`)
	}

	switch args[0] {
	case "validate":
		return runChannelValidate(args[1:])
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

func runChannelDeploy(args []string) error {
	fs := flag.NewFlagSet("deploy", flag.ExitOnError)
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega channel deploy <path>")
	}

	store, err := openChannelStore()
	if err != nil {
		return fmt.Errorf("open channel store: %w", err)
	}

	result, err := channel.Deploy(fs.Arg(0), store)
	if err != nil {
		return fmt.Errorf("deploy failed: %w", err)
	}

	fmt.Printf("Deployed %s revision %d hash %s\n", result.Name, result.Revision, result.Hash)
	return nil
}

func runChannelDiff(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega channel diff <path>")
	}

	store, err := openChannelStore()
	if err != nil {
		return fmt.Errorf("open channel store: %w", err)
	}

	result, err := channel.DiffLocal(fs.Arg(0), store)
	if err != nil {
		return fmt.Errorf("diff failed: %w", err)
	}

	if result.Identical {
		fmt.Println("No changes")
	} else {
		if result.DeployedHash == "" {
			fmt.Printf("Channel %s has no deployed version\n", result.ChannelName)
		} else {
			fmt.Printf("Local hash:    %s\n", result.LocalHash)
			fmt.Printf("Deployed hash: %s\n", result.DeployedHash)
			fmt.Printf("Channel %s has changes\n", result.ChannelName)
		}
	}
	return nil
}

func runChannelRollback(args []string) error {
	fs := flag.NewFlagSet("rollback", flag.ExitOnError)
	toHash := fs.String("to", "", "hash to roll back to")
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega channel rollback <name> [--to <hash>]")
	}

	store, err := openChannelStore()
	if err != nil {
		return fmt.Errorf("open channel store: %w", err)
	}

	name := fs.Arg(0)
	if err := channel.Rollback(name, *toHash, store); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Printf("Rolled back channel %s\n", name)
	return nil
}

func openChannelStore() (channelstore.ChannelStore, error) {
	dsn := os.Getenv("GHEGA_DATABASE_URL")
	if dsn == "" {
		dsn = "ghega.db"
	}
	return channelstore.NewSQLiteStore(dsn)
}
