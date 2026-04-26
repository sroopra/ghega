package config

import (
	"os"
	"strconv"
)

// Port returns the server port from the environment or a default.
func Port(defaultPort int) int {
	if p := os.Getenv("GHEGA_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			return v
		}
	}
	return defaultPort
}
