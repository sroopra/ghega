// Package messagestore provides persistent storage for message metadata
// and payload references. Raw payload bytes are never logged; only
// metadata and PayloadRef values are used for diagnostics.
package messagestore

import (
	"context"

	"github.com/sroopra/ghega/pkg/payloadref"
)

// Store persists message metadata and payload references.
// Implementations must never log raw payload bytes.
type Store interface {
	// Save persists the message metadata and payload reference.
	// The rawPayload is stored keyed by envelope.Ref.StorageID.
	Save(ctx context.Context, envelope *payloadref.Envelope, rawPayload []byte) error

	// GetMetadata retrieves message metadata by message ID.
	GetMetadata(ctx context.Context, messageID string) (*payloadref.Envelope, error)

	// ListByChannel returns message metadata for a channel, paginated.
	ListByChannel(ctx context.Context, channelID string, limit, offset int) ([]*payloadref.Envelope, error)
}

// ErrNotFound is returned when a message ID does not exist in the store.
type ErrNotFound struct {
	MessageID string
}

func (e *ErrNotFound) Error() string {
	return "message not found: " + e.MessageID
}
