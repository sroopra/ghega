package config

import (
	"errors"
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

// MigrationsDir returns the migrations directory from the environment or a default.
func MigrationsDir(defaultDir string) string {
	if d := os.Getenv("GHEGA_MIGRATIONS_DIR"); d != "" {
		return d
	}
	return defaultDir
}

// OIDCIssuer returns the OIDC provider issuer URL from the environment.
func OIDCIssuer() string {
	return os.Getenv("GHEGA_OIDC_ISSUER")
}

// OIDCClientID returns the OIDC client ID from the environment.
func OIDCClientID() string {
	return os.Getenv("GHEGA_OIDC_CLIENT_ID")
}

// OIDCClientSecret returns the OIDC client secret from the environment.
func OIDCClientSecret() string {
	return os.Getenv("GHEGA_OIDC_CLIENT_SECRET")
}

// OIDCRedirectURL returns the OIDC redirect URL from the environment or a default.
func OIDCRedirectURL() string {
	if u := os.Getenv("GHEGA_OIDC_REDIRECT_URL"); u != "" {
		return u
	}
	return "http://localhost:8080/auth/callback"
}

// SessionSecret returns the session cookie signing secret from the environment.
func SessionSecret() string {
	return os.Getenv("GHEGA_SESSION_SECRET")
}

// AuthEnabled returns whether authentication is enabled from the environment or a default.
func AuthEnabled() bool {
	if v := os.Getenv("GHEGA_AUTH_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return false
}

// AuthConfig bundles all authentication-related configuration.
type AuthConfig struct {
	Enabled       bool
	OIDCIssuer    string
	OIDCClientID  string
	OIDCClientSecret string
	OIDCRedirectURL  string
	SessionSecret string
}

// AuthConfigFromEnv builds an AuthConfig from environment variables.
func AuthConfigFromEnv() AuthConfig {
	return AuthConfig{
		Enabled:          AuthEnabled(),
		OIDCIssuer:       OIDCIssuer(),
		OIDCClientID:     OIDCClientID(),
		OIDCClientSecret: OIDCClientSecret(),
		OIDCRedirectURL:  OIDCRedirectURL(),
		SessionSecret:    SessionSecret(),
	}
}

// Validate returns an error if authentication is enabled but required fields are missing.
func (c AuthConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.OIDCIssuer == "" {
		return errors.New("GHEGA_OIDC_ISSUER is required when auth is enabled")
	}
	if c.OIDCClientID == "" {
		return errors.New("GHEGA_OIDC_CLIENT_ID is required when auth is enabled")
	}
	if c.OIDCClientSecret == "" {
		return errors.New("GHEGA_OIDC_CLIENT_SECRET is required when auth is enabled")
	}
	if c.SessionSecret == "" {
		return errors.New("GHEGA_SESSION_SECRET is required when auth is enabled")
	}
	return nil
}
