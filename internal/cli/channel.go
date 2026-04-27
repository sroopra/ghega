package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/sroopra/ghega/pkg/channel"
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
