package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler returns an http.Handler with all API routes registered.
func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/messages/", handleMessages)
	return mux
}

func handleMessages(w http.ResponseWriter, r *http.Request) {
	// Path pattern: /api/v1/messages/{id}/{action}
	// We need to extract the trailing action from the path.
	path := r.URL.Path
	if len(path) < len("/api/v1/messages/") {
		http.NotFound(w, r)
		return
	}

	// Check method
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simple path parsing: look for redeliver or replay suffix
	if strings.HasSuffix(path, "/redeliver") || strings.HasSuffix(path, "/replay") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not yet implemented"})
		return
	}

	http.NotFound(w, r)
}
