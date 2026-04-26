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

// MLLPHost returns the MLLP listener host from the environment or a default.
func MLLPHost(defaultHost string) string {
	if h := os.Getenv("GHEGA_MLLP_HOST"); h != "" {
		return h
	}
	return defaultHost
}

// MLLPPort returns the MLLP listener port from the environment or a default.
func MLLPPort(defaultPort int) int {
	if p := os.Getenv("GHEGA_MLLP_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			return v
		}
	}
	return defaultPort
}
