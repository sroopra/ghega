package cli

import (
	"fmt"
	"os"
)

func runGenerate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ghega generate <subcommand>")
	}

	switch args[0] {
	case "channel":
		fmt.Fprintln(os.Stderr, "not yet implemented")
		os.Exit(1)
		return nil
	default:
		return fmt.Errorf("unknown generate subcommand: %s", args[0])
	}
}
