package cli

import (
	"fmt"
)

// Execute parses and runs the CLI command.
func Execute(args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}

	switch args[0] {
	case "version", "--version", "-v":
		return runVersion()
	case "serve":
		return runServe(args[1:])
	case "channel":
		return runChannel(args[1:])
	case "message":
		return runMessage(args[1:])
	case "generate":
		return runGenerate(args[1:])
	case "--help", "-h", "help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func printUsage() {
	fmt.Println(`Ghega is an open-source healthcare integration engine.

Usage:
  ghega <command> [flags]

Commands:
  version            Print version information
  serve              Start the HTTP server
  channel validate   Validate a channel definition
  channel test       Run channel test fixtures
  channel deploy     Deploy a channel definition
  channel diff       Compare a local channel with the deployed version
  channel rollback   Rollback a channel to a previous revision
  message            Message management commands
  generate           Generate artifacts

Use "ghega <command> --help" for more information about a command.`)
}
