package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

const csrfCookieName = "__Host-csrf"
const csrfHeaderName = "X-CSRF-Token"

func generateCSRFToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func csrfTokenFromRequest(r *http.Request) string {
	return r.Header.Get(csrfHeaderName)
}

// CSRFMiddleware implements double-submit cookie CSRF protection.
// It generates or reuses a CSRF token, sets it as a cookie, and exposes it
// in the response header. Mutating requests under /api/v1/ must include a
// matching X-CSRF-Token header.
func (s *Server) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF processing for OIDC callback and health checks.
		if r.URL.Path == "/auth/callback" || r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		var token string
		cookie, err := r.Cookie(csrfCookieName)
		if err == nil && cookie.Value != "" {
			token = cookie.Value
		} else {
			token = generateCSRFToken()
			if token == "" {
				writeJSONError(w, http.StatusInternalServerError, "failed to generate csrf token")
				return
			}
		}

		http.SetCookie(w, &http.Cookie{
			Name:     csrfCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: false,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
		})

		w.Header().Set(csrfHeaderName, token)

		if isMutatingMethod(r.Method) {
			headerToken := csrfTokenFromRequest(r)
			if headerToken == "" || subtle.ConstantTimeCompare([]byte(headerToken), []byte(token)) != 1 {
				writeJSONError(w, http.StatusForbidden, "csrf token mismatch")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		return true
	}
	return false
}
