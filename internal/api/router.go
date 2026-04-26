package api

import (
	"net/http"
	"strings"

	"github.com/sroopra/ghega/pkg/messagestore"
)

// NewRouter creates an HTTP router with all API routes wired up.
// If devAuth is true, the auth middleware is bypassed.
func NewRouter(store messagestore.Store, registry ChannelRegistry, devAuth bool) http.Handler {
	server := NewServer(store, registry)
	auth := AuthMiddleware(devAuth)

	mux := http.NewServeMux()

	// Health check — no auth required.
	mux.HandleFunc("/healthz", server.Healthz)

	// API v1 routes — auth required.
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/v1/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		server.ListMessages(w, r)
	})
	apiMux.HandleFunc("/api/v1/messages/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		server.GetMessage(w, r)
	})
	apiMux.HandleFunc("/api/v1/channels", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		server.ListChannels(w, r)
	})
	apiMux.HandleFunc("/api/v1/channels/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		server.GetChannel(w, r)
	})

	// Wrap the API mux with auth middleware.
	mux.Handle("/api/", auth(apiMux))

	return mux
}

// stripPrefix strips the given prefix from the request URL path.
func stripPrefix(prefix string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		h.ServeHTTP(w, r)
	})
}
