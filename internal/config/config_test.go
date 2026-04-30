package config

import (
	"os"
	"testing"
)

func TestAuthEnabled(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{"default false", "", false},
		{"explicit false", "false", false},
		{"explicit true", "true", true},
		{"uppercase true", "TRUE", true},
		{"one true", "1", true},
		{"zero false", "0", false},
		{"invalid defaults false", "maybe", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GHEGA_AUTH_ENABLED", tt.envValue)
			if got := AuthEnabled(); got != tt.want {
				t.Errorf("AuthEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOIDCIssuer(t *testing.T) {
	t.Setenv("GHEGA_OIDC_ISSUER", "https://accounts.google.com")
	if got := OIDCIssuer(); got != "https://accounts.google.com" {
		t.Errorf("OIDCIssuer() = %q, want %q", got, "https://accounts.google.com")
	}
}

func TestOIDCClientID(t *testing.T) {
	t.Setenv("GHEGA_OIDC_CLIENT_ID", "my-client-id")
	if got := OIDCClientID(); got != "my-client-id" {
		t.Errorf("OIDCClientID() = %q, want %q", got, "my-client-id")
	}
}

func TestOIDCClientSecret(t *testing.T) {
	t.Setenv("GHEGA_OIDC_CLIENT_SECRET", "shh")
	if got := OIDCClientSecret(); got != "shh" {
		t.Errorf("OIDCClientSecret() = %q, want %q", got, "shh")
	}
}

func TestSessionSecret(t *testing.T) {
	t.Setenv("GHEGA_SESSION_SECRET", "super-secret")
	if got := SessionSecret(); got != "super-secret" {
		t.Errorf("SessionSecret() = %q, want %q", got, "super-secret")
	}
}

func TestOIDCRedirectURL(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     string
	}{
		{"default", "", "http://localhost:8080/auth/callback"},
		{"custom", "http://example.com/callback", "http://example.com/callback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GHEGA_OIDC_REDIRECT_URL", tt.envValue)
			if got := OIDCRedirectURL(); got != tt.want {
				t.Errorf("OIDCRedirectURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAuthConfigFromEnv(t *testing.T) {
	t.Setenv("GHEGA_AUTH_ENABLED", "true")
	t.Setenv("GHEGA_OIDC_ISSUER", "https://issuer.example.com")
	t.Setenv("GHEGA_OIDC_CLIENT_ID", "client-id")
	t.Setenv("GHEGA_OIDC_CLIENT_SECRET", "client-secret")
	t.Setenv("GHEGA_OIDC_REDIRECT_URL", "https://app.example.com/callback")
	t.Setenv("GHEGA_SESSION_SECRET", "session-secret")

	cfg := AuthConfigFromEnv()
	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.OIDCIssuer != "https://issuer.example.com" {
		t.Errorf("OIDCIssuer = %q, want %q", cfg.OIDCIssuer, "https://issuer.example.com")
	}
	if cfg.OIDCClientID != "client-id" {
		t.Errorf("OIDCClientID = %q, want %q", cfg.OIDCClientID, "client-id")
	}
	if cfg.OIDCClientSecret != "client-secret" {
		t.Errorf("OIDCClientSecret = %q, want %q", cfg.OIDCClientSecret, "client-secret")
	}
	if cfg.OIDCRedirectURL != "https://app.example.com/callback" {
		t.Errorf("OIDCRedirectURL = %q, want %q", cfg.OIDCRedirectURL, "https://app.example.com/callback")
	}
	if cfg.SessionSecret != "session-secret" {
		t.Errorf("SessionSecret = %q, want %q", cfg.SessionSecret, "session-secret")
	}
}

func TestAuthConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     AuthConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "disabled no fields",
			cfg:     AuthConfig{Enabled: false},
			wantErr: false,
		},
		{
			name: "enabled all fields",
			cfg: AuthConfig{
				Enabled:          true,
				OIDCIssuer:       "https://issuer.example.com",
				OIDCClientID:     "id",
				OIDCClientSecret: "secret",
				SessionSecret:    "session",
			},
			wantErr: false,
		},
		{
			name: "enabled missing issuer",
			cfg: AuthConfig{
				Enabled:          true,
				OIDCClientID:     "id",
				OIDCClientSecret: "secret",
				SessionSecret:    "session",
			},
			wantErr: true,
			errMsg:  "GHEGA_OIDC_ISSUER is required when auth is enabled",
		},
		{
			name: "enabled missing client id",
			cfg: AuthConfig{
				Enabled:          true,
				OIDCIssuer:       "https://issuer.example.com",
				OIDCClientSecret: "secret",
				SessionSecret:    "session",
			},
			wantErr: true,
			errMsg:  "GHEGA_OIDC_CLIENT_ID is required when auth is enabled",
		},
		{
			name: "enabled missing client secret",
			cfg: AuthConfig{
				Enabled:       true,
				OIDCIssuer:    "https://issuer.example.com",
				OIDCClientID:  "id",
				SessionSecret: "session",
			},
			wantErr: true,
			errMsg:  "GHEGA_OIDC_CLIENT_SECRET is required when auth is enabled",
		},
		{
			name: "enabled missing session secret",
			cfg: AuthConfig{
				Enabled:          true,
				OIDCIssuer:       "https://issuer.example.com",
				OIDCClientID:     "id",
				OIDCClientSecret: "secret",
			},
			wantErr: true,
			errMsg:  "GHEGA_SESSION_SECRET is required when auth is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Error() != tt.errMsg {
					t.Errorf("error message = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Ensure env cleanup between table-driven tests that might not use t.Setenv.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
