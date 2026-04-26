// Package payloadref provides types for referencing payloads without holding
// payload bytes in memory. This ensures PHI never leaks through string
// formatting or logging.
package payloadref

import (
	"fmt"
	"time"
)

// PayloadRef references a payload by storage ID/location.
// It must NOT hold payload bytes.
type PayloadRef struct {
	StorageID string
	Location  string
}

// String returns a PHI-safe representation that contains only the
// storage metadata. Payload bytes never appear.
func (p PayloadRef) String() string {
	return fmt.Sprintf("PayloadRef{StorageID:%s Location:%s}", p.StorageID, p.Location)
}

// GoString implements fmt.GoStringer to prevent %#v from revealing fields
// that could be modified in the future to contain sensitive data.
func (p PayloadRef) GoString() string {
	return p.String()
}

// Envelope contains message metadata and a reference to the payload.
// It never holds payload bytes directly.
type Envelope struct {
	ChannelID  string
	MessageID  string
	ReceivedAt time.Time
	Ref        PayloadRef
}

// String returns a PHI-safe representation that contains only metadata IDs
// and timestamps. Payload bytes never appear.
func (e Envelope) String() string {
	return fmt.Sprintf(
		"Envelope{ChannelID:%s MessageID:%s ReceivedAt:%s Ref:%s}",
		e.ChannelID,
		e.MessageID,
		e.ReceivedAt.Format(time.RFC3339Nano),
		e.Ref.String(),
	)
}

// GoString implements fmt.GoStringer to prevent %#v from revealing fields
// that could be modified in the future to contain sensitive data.
func (e Envelope) GoString() string {
	return e.String()
}
