// Package channelstore provides persistent storage for Ghega channel revisions
// and deployment audit records.
package channelstore

import (
	"context"
	"time"
)

// ChannelStore persists channel YAML definitions, their revision history, and
// deployment audit trails.
type ChannelStore interface {
	// SaveChannel persists a channel revision. If revision is <= 0 it is auto-incremented.
	SaveChannel(ctx context.Context, name, hash string, yamlBytes []byte, revision int) error

	// GetChannel returns the latest revision for the given channel name.
	GetChannel(ctx context.Context, name string) (*ChannelRecord, error)

	// GetChannelRevision returns a specific revision by its hash.
	GetChannelRevision(ctx context.Context, name, hash string) (*ChannelRecord, error)

	// ListChannelRevisions returns all revisions for a channel ordered by revision desc.
	ListChannelRevisions(ctx context.Context, name string) ([]ChannelRecord, error)

	// RollbackChannel verifies the hash exists and records a rollback audit entry.
	RollbackChannel(ctx context.Context, name, hash string) error

	// SaveDeploymentAudit records a deployment action (deploy, rollback, etc.).
	SaveDeploymentAudit(ctx context.Context, channelName, hash, action string) error

	// ListDeploymentAudit returns all audit entries for a channel ordered by timestamp desc.
	ListDeploymentAudit(ctx context.Context, channelName string) ([]AuditRecord, error)
}

// ChannelRecord represents a single persisted channel revision.
type ChannelRecord struct {
	Name       string
	Hash       string
	YAML       []byte
	Revision   int
	DeployedAt time.Time
}

// AuditRecord represents a single deployment audit entry.
type AuditRecord struct {
	ChannelName string
	Hash        string
	Action      string
	Timestamp   time.Time
}

// ErrNotFound is returned when a channel or revision does not exist.
type ErrNotFound struct {
	Name string
	Hash string
}

func (e *ErrNotFound) Error() string {
	if e.Hash != "" {
		return "channel revision not found: " + e.Name + "@" + e.Hash
	}
	return "channel not found: " + e.Name
}
