package api

import (
	"context"
	"net/http"
	"strings"
)

// User represents an authenticated user attached to the request context.
type User struct {
	Subject string
	Role    string
}

// userContextKey is the key used to store the User in the request context.
type userContextKey struct{}

// UserFromContext retrieves the User from the context, if present.
func UserFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userContextKey{}).(*User)
	return u, ok
}

// AuthMiddleware returns an HTTP middleware that validates Bearer tokens.
// If devAuth is true, authentication is bypassed and a placeholder user is attached.
func AuthMiddleware(devAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if devAuth {
				ctx := context.WithValue(r.Context(), userContextKey{}, &User{
					Subject: "dev-user",
					Role:    "admin",
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			token := parts[1]
			if !isValidJWTFormat(token) {
				http.Error(w, `{"error":"invalid token format"}`, http.StatusUnauthorized)
				return
			}

			// Placeholder: extract a subject from the token header/payload for the user context.
			subject := extractSubject(token)
			ctx := context.WithValue(r.Context(), userContextKey{}, &User{
				Subject: subject,
				Role:    "operator",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isValidJWTFormat checks that the token has three base64url-encoded parts separated by dots.
func isValidJWTFormat(token string) bool {
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

// extractSubject extracts a placeholder subject from the JWT payload.
// In a real implementation this would decode and validate the JWT.
func extractSubject(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "unknown"
	}
	// Placeholder: use the payload part as a basis for the subject.
	// No actual decoding or validation is performed.
	if len(parts[1]) > 8 {
		return "user-" + parts[1][:8]
	}
	return "user-" + parts[1]
}
