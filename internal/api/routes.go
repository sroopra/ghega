package api

import (
	"net/http"
)

// NewRouter builds the API router with all routes and the auth middleware applied.
func NewRouter(handler *Handler, devAuth bool) http.Handler {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /healthz", handler.Healthz)

	// Protected API routes
	auth := AuthMiddleware(devAuth)

	mux.Handle("GET /api/v1/messages", auth(http.HandlerFunc(handler.ListMessages)))
	mux.Handle("GET /api/v1/messages/{id}", auth(http.HandlerFunc(handler.GetMessage)))
	mux.Handle("GET /api/v1/channels", auth(http.HandlerFunc(handler.ListChannels)))
	mux.Handle("GET /api/v1/channels/{id}", auth(http.HandlerFunc(handler.GetChannel)))

	return mux
}
