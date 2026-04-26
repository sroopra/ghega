// Package api provides the REST API handlers and BFF auth middleware for Ghega.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/payloadref"
)

// ChannelConfig represents a channel configuration.
type ChannelConfig struct {
	Name        string            `json:"name" yaml:"name"`
	Source      SourceConfig      `json:"source" yaml:"source"`
	Destination DestinationConfig `json:"destination" yaml:"destination"`
	Mapping     MappingConfig     `json:"mapping" yaml:"mapping"`
}

// SourceConfig represents the source configuration for a channel.
type SourceConfig struct {
	Type string `json:"type" yaml:"type"`
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
	Port int    `json:"port,omitempty" yaml:"port,omitempty"`
}

// DestinationConfig represents the destination configuration for a channel.
type DestinationConfig struct {
	Type string `json:"type" yaml:"type"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

// MappingConfig represents the mapping configuration for a channel.
type MappingConfig struct {
	MessageType string `json:"messageType" yaml:"messageType"`
}

// MessageResponse is the JSON representation of message metadata.
// It never includes payload bytes.
type MessageResponse struct {
	ChannelID  string    `json:"channel_id"`
	MessageID  string    `json:"message_id"`
	ReceivedAt time.Time `json:"received_at"`
	Status     string    `json:"status"`
	StorageID  string    `json:"storage_id"`
	Location   string    `json:"location"`
}

// PaginatedMessagesResponse is the JSON response for listing messages.
type PaginatedMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// envelopeToResponse converts a payloadref.Envelope to a MessageResponse.
func envelopeToResponse(env *payloadref.Envelope) MessageResponse {
	return MessageResponse{
		ChannelID:  env.ChannelID,
		MessageID:  env.MessageID,
		ReceivedAt: env.ReceivedAt,
		Status:     env.Status,
		StorageID:  env.Ref.StorageID,
		Location:   env.Ref.Location,
	}
}

// Server holds the API dependencies.
type Server struct {
	MessageStore    messagestore.Store
	ChannelRegistry ChannelRegistry
}

// ChannelRegistry provides access to channel configurations.
type ChannelRegistry interface {
	List() []*ChannelConfig
	Get(name string) (*ChannelConfig, bool)
}

// NewServer creates a new API server with the given dependencies.
func NewServer(store messagestore.Store, registry ChannelRegistry) *Server {
	return &Server{
		MessageStore:    store,
		ChannelRegistry: registry,
	}
}

// Healthz handles GET /healthz.
func (s *Server) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ListMessages handles GET /api/v1/messages.
func (s *Server) ListMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	envelopes, err := s.MessageStore.List(ctx, limit, offset)
	if err != nil {
		http.Error(w, `{"error":"failed to list messages"}`, http.StatusInternalServerError)
		return
	}

	msgs := make([]MessageResponse, 0, len(envelopes))
	for _, env := range envelopes {
		msgs = append(msgs, envelopeToResponse(env))
	}

	resp := PaginatedMessagesResponse{
		Messages: msgs,
		Limit:    limit,
		Offset:   offset,
		Total:    len(msgs),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetMessage handles GET /api/v1/messages/:id.
func (s *Server) GetMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/messages/")
	if id == "" {
		http.Error(w, `{"error":"message id is required"}`, http.StatusBadRequest)
		return
	}

	env, err := s.MessageStore.GetMetadata(ctx, id)
	if err != nil {
		if _, ok := err.(*messagestore.ErrNotFound); ok {
			http.Error(w, `{"error":"message not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"failed to get message"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(envelopeToResponse(env))
}

// ListChannels handles GET /api/v1/channels.
func (s *Server) ListChannels(w http.ResponseWriter, r *http.Request) {
	channels := s.ChannelRegistry.List()
	resp := make([]*ChannelConfig, 0, len(channels))
	for _, ch := range channels {
		resp = append(resp, ch)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetChannel handles GET /api/v1/channels/:id.
func (s *Server) GetChannel(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/channels/")
	if id == "" {
		http.Error(w, `{"error":"channel id is required"}`, http.StatusBadRequest)
		return
	}

	ch, ok := s.ChannelRegistry.Get(id)
	if !ok {
		http.Error(w, `{"error":"channel not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ch)
}

// InMemoryChannelRegistry is a simple in-memory channel registry.
type InMemoryChannelRegistry struct {
	channels map[string]*ChannelConfig
}

// NewInMemoryChannelRegistry creates a new empty InMemoryChannelRegistry.
func NewInMemoryChannelRegistry() *InMemoryChannelRegistry {
	return &InMemoryChannelRegistry{
		channels: make(map[string]*ChannelConfig),
	}
}

// Register adds a channel config to the registry.
func (r *InMemoryChannelRegistry) Register(ch *ChannelConfig) {
	r.channels[ch.Name] = ch
}

// List returns all registered channel configs.
func (r *InMemoryChannelRegistry) List() []*ChannelConfig {
	out := make([]*ChannelConfig, 0, len(r.channels))
	for _, ch := range r.channels {
		out = append(out, ch)
	}
	return out
}

// Get retrieves a channel config by name.
func (r *InMemoryChannelRegistry) Get(name string) (*ChannelConfig, bool) {
	ch, ok := r.channels[name]
	return ch, ok
}

// Ensure InMemoryChannelRegistry implements ChannelRegistry.
var _ ChannelRegistry = (*InMemoryChannelRegistry)(nil)
