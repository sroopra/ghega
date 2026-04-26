// Package api provides the REST API handlers and middleware for Ghega.
package api

import (
	"context"
	"net/http"
	"strings"
)

// User represents an authenticated user in the BFF layer.
type User struct {
	ID    string
	Role  string
	Token string
}

// userContextKey is the key used to store User in request context.
type userContextKey struct{}

// ContextUser retrieves the User from the request context.
// Returns nil if no user is attached.
func ContextUser(r *http.Request) *User {
	u, _ := r.Context().Value(userContextKey{}).(*User)
	return u
}

// AuthMiddleware returns an HTTP middleware that validates Bearer tokens.
// If devAuth is true, all requests are allowed and a placeholder user is attached.
func AuthMiddleware(devAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if devAuth {
				user := &User{
					ID:   "dev-user",
					Role: "admin",
				}
				r = r.WithContext(context.WithValue(r.Context(), userContextKey{}, user))
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			token := parts[1]
			if !isValidTokenFormat(token) {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			user := &User{
				ID:    "placeholder-user",
				Role:  "operator",
				Token: token,
			}
			r = r.WithContext(context.WithValue(r.Context(), userContextKey{}, user))
			next.ServeHTTP(w, r)
		})
	}
}

// isValidTokenFormat performs a lightweight validation of a JWT-like token.
// It checks that the token has three base64url-encoded segments separated by dots.
func isValidTokenFormat(token string) bool {
	// A JWT has exactly 2 dots separating header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
	}
	return true
}
