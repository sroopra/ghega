package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sroopra/ghega/internal/config"
	"github.com/sroopra/ghega/internal/session"
	"golang.org/x/oauth2"
)

const stateCookieName = "ghega_oidc_state"

// OIDCProvider holds OIDC discovery, OAuth2 configuration, and session management.
type OIDCProvider struct {
	provider *oidc.Provider
	config   oauth2.Config
	mgr      *session.Manager
	verifier *oidc.IDTokenVerifier
}

// NewOIDCProvider performs OIDC discovery and returns a configured OIDCProvider.
func NewOIDCProvider(ctx context.Context, cfg config.AuthConfig, mgr *session.Manager) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(ctx, cfg.OIDCIssuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}

	oidcConfig := &oidc.Config{
		ClientID: cfg.OIDCClientID,
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		RedirectURL:  cfg.OIDCRedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &OIDCProvider{
		provider: provider,
		config:   oauth2Config,
		mgr:      mgr,
		verifier: provider.Verifier(oidcConfig),
	}, nil
}

// HandleLogin generates a random state, stores it in a short-lived cookie, and redirects to the OIDC provider.
func (op *OIDCProvider) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := randomState()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate state")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, op.config.AuthCodeURL(state), http.StatusFound)
}

// HandleCallback validates the state cookie, exchanges the code for a token, verifies the ID token,
// creates a session, and redirects to /.
func (op *OIDCProvider) HandleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "missing state cookie")
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	if r.URL.Query().Get("state") != stateCookie.Value {
		writeJSONError(w, http.StatusBadRequest, "invalid state")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeJSONError(w, http.StatusBadRequest, "missing code")
		return
	}

	ctx := r.Context()
	oauth2Token, err := op.config.Exchange(ctx, code)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "failed to exchange token")
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "no id_token in token response")
		return
	}

	idToken, err := op.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid id_token")
		return
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to parse claims")
		return
	}

	if claims.Email == "" {
		claims.Email = idToken.Subject
	}

	op.mgr.CreateSession(w, claims.Email, claims.Name, []string{"user"})
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleLogout destroys the session and redirects to /.
func (op *OIDCProvider) HandleLogout(w http.ResponseWriter, r *http.Request) {
	op.mgr.DestroySession(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

// handleMe returns the current user's session information.
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	sess := session.SessionFromContext(r.Context())
	if sess == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, map[string]any{
		"email": sess.Email,
		"name":  sess.Name,
		"roles": sess.Roles,
	})
}

// AuthMiddleware authenticates requests. When auth is disabled, it injects a dev session.
// When auth is enabled, it requires a valid session cookie.
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.authConfig.Enabled {
			devSess := &session.Session{
				ID:        "dev-session",
				Email:     "dev@ghega.local",
				Name:      "Developer",
				Roles:     []string{"dev"},
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
			}
			ctx := session.ContextWithSession(r.Context(), devSess)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		sess := s.sessionMgr.ReadSession(r)
		if sess == nil {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := session.ContextWithSession(r.Context(), sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORSMiddleware adds CORS headers. When auth is enabled, it allows credentials and does not use a wildcard origin.
func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.authConfig.Enabled {
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
