package server

import (
	"log/slog"
	"net/http"
)

// AuthMiddleware is a placeholder authentication middleware for the Ghega BFF.
// It logs all incoming requests (method, path, remote address) and rejects
// requests that carry an Authorization header with the literal value
// "Bearer invalid" with HTTP 401.
//
// TODO: Replace this placeholder with real authentication (JWT/OAuth2)
// before any production deployment.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("bff request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		if auth := r.Header.Get("Authorization"); auth == "Bearer invalid" {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware adds CORS headers for local development.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
