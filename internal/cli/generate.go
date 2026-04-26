package cli

import (
	"fmt"

	"github.com/sroopra/ghega/internal/cli/generate"
)

func runGenerate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ghega generate <subcommand>")
	}

	switch args[0] {
	case "channel":
		return generate.RunChannelGenerate(args[1:])
	default:
		return fmt.Errorf("unknown generate subcommand: %s", args[0])
	}
}
