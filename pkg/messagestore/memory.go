package messagestore

import (
	"context"
	"sync"

	"github.com/sroopra/ghega/pkg/payloadref"
)

// InMemoryStore is a development/testing implementation of Store.
// It stores raw payloads in a map keyed by storage ID and metadata
// indexed by message ID.
type InMemoryStore struct {
	mu       sync.RWMutex
	metadata map[string]*payloadref.Envelope // keyed by message ID
	payloads map[string][]byte               // keyed by storage ID
}

// NewInMemoryStore creates a new empty InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		metadata: make(map[string]*payloadref.Envelope),
		payloads: make(map[string][]byte),
	}
}

// Save persists the message metadata and raw payload.
// Raw payload bytes are stored in a map keyed by envelope.Ref.StorageID.
func (s *InMemoryStore) Save(_ context.Context, envelope *payloadref.Envelope, rawPayload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := *envelope
	s.metadata[envelope.MessageID] = &cp
	s.payloads[envelope.Ref.StorageID] = append([]byte(nil), rawPayload...)
	return nil
}

// GetMetadata retrieves message metadata by message ID.
func (s *InMemoryStore) GetMetadata(_ context.Context, messageID string) (*payloadref.Envelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	env, ok := s.metadata[messageID]
	if !ok {
		return nil, &ErrNotFound{MessageID: messageID}
	}
	cp := *env
	return &cp, nil
}

// ListByChannel returns message metadata for a channel, paginated.
func (s *InMemoryStore) ListByChannel(_ context.Context, channelID string, limit, offset int) ([]*payloadref.Envelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matched []*payloadref.Envelope
	for _, env := range s.metadata {
		if env.ChannelID == channelID {
			matched = append(matched, env)
		}
	}

	if offset >= len(matched) {
		return []*payloadref.Envelope{}, nil
	}
	end := offset + limit
	if end > len(matched) || limit <= 0 {
		end = len(matched)
	}

	out := make([]*payloadref.Envelope, 0, end-offset)
	for i := offset; i < end; i++ {
		cp := *matched[i]
		out = append(out, &cp)
	}
	return out, nil
}

// List returns all message metadata, paginated.
func (s *InMemoryStore) List(_ context.Context, limit, offset int) ([]*payloadref.Envelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matched []*payloadref.Envelope
	for _, env := range s.metadata {
		matched = append(matched, env)
	}

	if offset >= len(matched) {
		return []*payloadref.Envelope{}, nil
	}
	end := offset + limit
	if end > len(matched) || limit <= 0 {
		end = len(matched)
	}

	out := make([]*payloadref.Envelope, 0, end-offset)
	for i := offset; i < end; i++ {
		cp := *matched[i]
		out = append(out, &cp)
	}
	return out, nil
}

// GetPayload retrieves raw payload bytes by storage ID.
// This is intended for testing and internal use only.
func (s *InMemoryStore) GetPayload(storageID string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.payloads[storageID]
	if !ok {
		return nil, false
	}
	return append([]byte(nil), p...), true
}
