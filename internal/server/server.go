// Package server implements the Ghega HTTP API and BFF middleware.
package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"
	"time"

	"github.com/sroopra/ghega"
	"github.com/sroopra/ghega/internal/alerts"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/payloadref"
)

// Server holds dependencies for the HTTP API.
type Server struct {
	store      messagestore.Store
	alertStore alerts.AlertStore
}

// New creates a new Server with the given store and alertStore.
func New(store messagestore.Store, alertStore alerts.AlertStore) *Server {
	return &Server{store: store, alertStore: alertStore}
}

// messageMetadataResponse is the JSON shape expected by the UI.
type messageMetadataResponse struct {
	ID         string `json:"id"`
	ChannelID  string `json:"channel_id"`
	MessageID  string `json:"message_id"`
	Status     string `json:"status"`
	ReceivedAt string `json:"received_at"`
	StorageID  string `json:"storage_id"`
	Location   string `json:"location"`
}

// channelResponse is the JSON shape for a channel.
type channelResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func envelopeToResponse(env *payloadref.Envelope) messageMetadataResponse {
	return messageMetadataResponse{
		ID:         env.MessageID,
		ChannelID:  env.ChannelID,
		MessageID:  env.MessageID,
		Status:     env.Status,
		ReceivedAt: env.ReceivedAt.Format(time.RFC3339Nano),
		StorageID:  env.Ref.StorageID,
		Location:   env.Ref.Location,
	}
}

// Handler returns the HTTP handler for the Ghega API.
func (s *Server) Handler() http.Handler {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /messages", s.handleListMessages)
	apiMux.HandleFunc("GET /messages/{id}", s.handleGetMessage)
	apiMux.HandleFunc("POST /messages/{id}/redeliver", s.handleRedeliver)
	apiMux.HandleFunc("POST /messages/{id}/replay", s.handleReplay)
	apiMux.HandleFunc("GET /channels", s.handleListChannels)
	apiMux.HandleFunc("GET /alerts", s.handleListAlerts)

	wrapped := CORSMiddleware(AuthMiddleware(apiMux))

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", wrapped))
	mux.HandleFunc("/healthz", s.handleHealthz)

	sub, err := fs.Sub(ghega.UIFS, "ui/dist")
	if err != nil {
		panic("ui/dist not embedded: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))
	mux.Handle("/", fileServer)

	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleListMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	channelID := r.URL.Query().Get("channel_id")

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	var envelopes []*payloadref.Envelope
	var err error
	if channelID != "" {
		envelopes, err = s.store.ListByChannel(ctx, channelID, limit, offset)
	} else {
		envelopes, err = s.store.ListAll(ctx, limit, offset)
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}

	resp := make([]messageMetadataResponse, len(envelopes))
	for i, env := range envelopes {
		resp[i] = envelopeToResponse(env)
	}
	writeJSON(w, resp)
}

func (s *Server) handleGetMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "missing message id")
		return
	}

	env, err := s.store.GetMetadata(ctx, id)
	if err != nil {
		if _, ok := err.(*messagestore.ErrNotFound); ok {
			writeJSONError(w, http.StatusNotFound, "message not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to get message")
		return
	}

	writeJSON(w, envelopeToResponse(env))
}

func (s *Server) handleListChannels(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, []channelResponse{
		{ID: "adt-a01", Name: "ADT A01 MLLP to HTTP"},
	})
}

func (s *Server) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	alertsList, err := s.alertStore.List()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}
	writeJSON(w, alertsList)
}

func (s *Server) handleRedeliver(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotImplemented, "not yet implemented")
}

func (s *Server) handleReplay(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotImplemented, "not yet implemented")
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
