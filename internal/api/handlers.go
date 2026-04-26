package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/payloadref"
)

// Channel represents a Ghega channel configuration.
type Channel struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Source      map[string]string `json:"source,omitempty"`
	Destination map[string]string `json:"destination,omitempty"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
}

// ChannelStore provides access to channel configurations.
type ChannelStore interface {
	List() ([]Channel, error)
	Get(id string) (Channel, bool)
}

// InMemoryChannelStore is a simple in-memory channel registry.
type InMemoryChannelStore struct {
	channels map[string]Channel
}

// NewInMemoryChannelStore creates an empty InMemoryChannelStore.
func NewInMemoryChannelStore() *InMemoryChannelStore {
	return &InMemoryChannelStore{channels: make(map[string]Channel)}
}

// Register adds a channel to the store.
func (s *InMemoryChannelStore) Register(ch Channel) {
	s.channels[ch.ID] = ch
}

// List returns all registered channels.
func (s *InMemoryChannelStore) List() ([]Channel, error) {
	out := make([]Channel, 0, len(s.channels))
	for _, ch := range s.channels {
		out = append(out, ch)
	}
	return out, nil
}

// Get retrieves a channel by ID.
func (s *InMemoryChannelStore) Get(id string) (Channel, bool) {
	ch, ok := s.channels[id]
	return ch, ok
}

// MessageMetadataResponse is the JSON representation of message metadata.
// It never includes payload bytes.
type MessageMetadataResponse struct {
	ChannelID  string    `json:"channel_id"`
	MessageID  string    `json:"message_id"`
	ReceivedAt time.Time `json:"received_at"`
	Status     string    `json:"status"`
	StorageID  string    `json:"storage_id"`
	Location   string    `json:"location"`
}

func envelopeToResponse(env *payloadref.Envelope) MessageMetadataResponse {
	return MessageMetadataResponse{
		ChannelID:  env.ChannelID,
		MessageID:  env.MessageID,
		ReceivedAt: env.ReceivedAt,
		Status:     env.Status,
		StorageID:  env.Ref.StorageID,
		Location:   env.Ref.Location,
	}
}

// Handler holds dependencies for the API handlers.
type Handler struct {
	MessageStore messagestore.Store
	ChannelStore ChannelStore
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(msgStore messagestore.Store, chStore ChannelStore) *Handler {
	return &Handler{
		MessageStore: msgStore,
		ChannelStore: chStore,
	}
}

// Healthz handles GET /healthz.
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ListMessages handles GET /api/v1/messages.
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	ctx := r.Context()
	var envelopes []*payloadref.Envelope
	var err error

	if channelID := r.URL.Query().Get("channel_id"); channelID != "" {
		envelopes, err = h.MessageStore.ListByChannel(ctx, channelID, limit, offset)
	} else {
		envelopes, err = h.MessageStore.ListAll(ctx, limit, offset)
	}

	if err != nil {
		http.Error(w, `{"error":"failed to list messages"}`, http.StatusInternalServerError)
		return
	}

	resp := make([]MessageMetadataResponse, 0, len(envelopes))
	for _, env := range envelopes {
		resp = append(resp, envelopeToResponse(env))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetMessage handles GET /api/v1/messages/:id.
func (h *Handler) GetMessage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"missing message id"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	env, err := h.MessageStore.GetMetadata(ctx, id)
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
func (h *Handler) ListChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := h.ChannelStore.List()
	if err != nil {
		http.Error(w, `{"error":"failed to list channels"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(channels)
}

// GetChannel handles GET /api/v1/channels/:id.
func (h *Handler) GetChannel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"missing channel id"}`, http.StatusBadRequest)
		return
	}

	ch, ok := h.ChannelStore.Get(id)
	if !ok {
		http.Error(w, `{"error":"channel not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ch)
}
