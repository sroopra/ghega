package channel

import (
	"context"
	"fmt"
	"os"

	"github.com/sroopra/ghega/pkg/channelstore"
)

// DeployResult holds the outcome of a channel deployment.
type DeployResult struct {
	Name         string
	Hash         string
	Revision     int
	PreviousHash string
}

// Deploy reads a channel YAML file, validates it, computes its hash, and
// persists it to the given store. If the hash already exists for the channel
// the operation is a no-op (idempotent).
func Deploy(channelPath string, store channelstore.ChannelStore) (*DeployResult, error) {
	data, err := os.ReadFile(channelPath)
	if err != nil {
		return nil, fmt.Errorf("read channel file: %w", err)
	}

	ch, valErrs := ValidateYAML(data)
	if ch != nil {
		valErrs = append(valErrs, ValidatePolicies(ch)...)
	}
	if len(valErrs) > 0 {
		return nil, fmt.Errorf("validation failed: %d errors", len(valErrs))
	}

	hash, err := HashChannel(ch)
	if err != nil {
		return nil, fmt.Errorf("hash channel: %w", err)
	}

	ctx := context.Background()

	var previousHash string
	current, err := store.GetChannel(ctx, ch.Name)
	if err == nil {
		previousHash = current.Hash
	} else if _, ok := err.(*channelstore.ErrNotFound); !ok {
		return nil, fmt.Errorf("get current channel: %w", err)
	}

	// Idempotency check: if the hash already exists anywhere in history, no-op.
	existing, err := store.GetChannelRevision(ctx, ch.Name, hash)
	if err == nil {
		return &DeployResult{
			Name:         ch.Name,
			Hash:         hash,
			Revision:     existing.Revision,
			PreviousHash: previousHash,
		}, nil
	}
	if _, ok := err.(*channelstore.ErrNotFound); !ok {
		return nil, fmt.Errorf("check existing revision: %w", err)
	}

	if err := store.SaveChannel(ctx, ch.Name, hash, data, 0); err != nil {
		return nil, fmt.Errorf("save channel: %w", err)
	}

	saved, err := store.GetChannel(ctx, ch.Name)
	if err != nil {
		return nil, fmt.Errorf("get saved channel: %w", err)
	}

	if err := store.SaveDeploymentAudit(ctx, ch.Name, hash, "deploy"); err != nil {
		return nil, fmt.Errorf("save deployment audit: %w", err)
	}

	return &DeployResult{
		Name:         ch.Name,
		Hash:         hash,
		Revision:     saved.Revision,
		PreviousHash: previousHash,
	}, nil
}
