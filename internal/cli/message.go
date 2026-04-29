package cli

import (
	"flag"
	"fmt"
)

func runMessage(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ghega message <subcommand> [flags]")
	}

	switch args[0] {
	case "redeliver":
		return runMessageRedeliver(args[1:])
	case "replay":
		return runMessageReplay(args[1:])
	case "replay-preview":
		return runMessageReplayPreview(args[1:])
	default:
		return fmt.Errorf("unknown message subcommand: %s", args[0])
	}
}

func runMessageRedeliver(args []string) error {
	fs := flag.NewFlagSet("redeliver", flag.ExitOnError)
	_ = fs.String("destination", "", "Destination endpoint")
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega message redeliver <message-id> --destination <dest>")
	}

	return fmt.Errorf("redeliver not yet implemented")
}

func runMessageReplay(args []string) error {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	_ = fs.Bool("as-new", false, "Replay as a new message")
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega message replay <message-id> --as-new")
	}

	return fmt.Errorf("replay not yet implemented")
}

func runMessageReplayPreview(args []string) error {
	fs := flag.NewFlagSet("replay-preview", flag.ExitOnError)
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega message replay-preview <message-id>")
	}

	return fmt.Errorf("replay-preview not yet implemented")
}
