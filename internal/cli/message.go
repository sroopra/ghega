package cli

import (
	"flag"
	"fmt"
	"os"
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

	fmt.Fprintln(os.Stderr, "not yet implemented")
	os.Exit(1)
	return nil
}

func runMessageReplay(args []string) error {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	_ = fs.Bool("as-new", false, "Replay as a new message")
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega message replay <message-id> --as-new")
	}

	fmt.Fprintln(os.Stderr, "not yet implemented")
	os.Exit(1)
	return nil
}

func runMessageReplayPreview(args []string) error {
	fs := flag.NewFlagSet("replay-preview", flag.ExitOnError)
	_ = fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: ghega message replay-preview <message-id>")
	}

	fmt.Fprintln(os.Stderr, "not yet implemented")
	os.Exit(1)
	return nil
}
