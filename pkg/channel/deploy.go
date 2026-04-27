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

// Deploy reads a channel YAML file, validates it, hashes it, and persists it
// to the given store. If the hash already exists for this channel the deploy
// is a no-op (idempotent).
func Deploy(channelPath string, store channelstore.ChannelStore) (*DeployResult, error) {
	data, err := os.ReadFile(channelPath)
	if err != nil {
		return nil, fmt.Errorf("read channel file: %w", err)
	}

	ch, verrs := ValidateYAML(data)
	if len(verrs) > 0 {
		return nil, fmt.Errorf("validation failed: %d errors", len(verrs))
	}

	hash, err := HashChannel(ch)
	if err != nil {
		return nil, fmt.Errorf("hash channel: %w", err)
	}

	ctx := context.Background()

	var previousHash string
	if prev, err := store.GetChannel(ctx, ch.Name); err == nil {
		previousHash = prev.Hash
	}

	// Idempotency: if this hash is already known, return existing info.
	if existing, err := store.GetChannelRevision(ctx, ch.Name, hash); err == nil {
		return &DeployResult{
			Name:         ch.Name,
			Hash:         hash,
			Revision:     existing.Revision,
			PreviousHash: previousHash,
		}, nil
	}

	if err := store.SaveChannel(ctx, ch.Name, hash, data, 0); err != nil {
		return nil, fmt.Errorf("save channel: %w", err)
	}

	if err := store.SaveDeploymentAudit(ctx, ch.Name, hash, "deploy"); err != nil {
		return nil, fmt.Errorf("save deployment audit: %w", err)
	}

	rec, err := store.GetChannel(ctx, ch.Name)
	if err != nil {
		return nil, fmt.Errorf("get channel after save: %w", err)
	}

	return &DeployResult{
		Name:         ch.Name,
		Hash:         hash,
		Revision:     rec.Revision,
		PreviousHash: previousHash,
	}, nil
}
