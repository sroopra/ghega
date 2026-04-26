package cli

import (
	"flag"
	"fmt"
	"os"
)

func runChannel(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ghega channel validate <path>")
	}

	switch args[0] {
	case "validate":
		return runChannelValidate(args[1:])
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

	fmt.Fprintln(os.Stderr, "not yet implemented")
	os.Exit(1)
	return nil
}
