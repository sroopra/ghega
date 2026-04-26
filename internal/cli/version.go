package cli

import (
	"fmt"
	"runtime"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func runVersion() error {
	fmt.Printf("Ghega version %s\n", version)
	fmt.Printf("  commit:  %s\n", commit)
	fmt.Printf("  go:      %s\n", runtime.Version())
	fmt.Printf("  built:   %s\n", buildDate)
	return nil
}
