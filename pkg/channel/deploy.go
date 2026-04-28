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
// persists it via the provided ChannelStore. If the hash already exists the
// deployment is a no-op.
func Deploy(channelPath string, store channelstore.ChannelStore) (*DeployResult, error) {
	ctx := context.Background()

	data, err := os.ReadFile(channelPath)
	if err != nil {
		return nil, fmt.Errorf("read channel file: %w", err)
	}

	ch, valErrs := ValidateYAML(data)
	if len(valErrs) > 0 {
		return nil, fmt.Errorf("validation failed: %v", valErrs)
	}

	polErrs := ValidatePolicies(ch)
	if len(polErrs) > 0 {
		return nil, fmt.Errorf("policy validation failed: %v", polErrs)
	}

	hash, err := HashChannel(ch)
	if err != nil {
		return nil, fmt.Errorf("hash channel: %w", err)
	}

	// Check if this exact revision already exists.
	if existing, err := store.GetChannelRevision(ctx, ch.Name, hash); err == nil {
		return &DeployResult{
			Name:     ch.Name,
			Hash:     hash,
			Revision: existing.Revision,
		}, nil
	}

	// Capture the previous deployed version before saving.
	var previousHash string
	if prev, err := store.GetChannel(ctx, ch.Name); err == nil {
		previousHash = prev.Hash
	}

	if err := store.SaveChannel(ctx, ch.Name, hash, data, 0); err != nil {
		return nil, fmt.Errorf("save channel: %w", err)
	}

	if err := store.SaveDeploymentAudit(ctx, ch.Name, hash, "deploy"); err != nil {
		return nil, fmt.Errorf("save deployment audit: %w", err)
	}

	// Retrieve the newly saved revision to return its revision number.
	rec, err := store.GetChannelRevision(ctx, ch.Name, hash)
	if err != nil {
		return nil, fmt.Errorf("get saved revision: %w", err)
	}

	return &DeployResult{
		Name:         ch.Name,
		Hash:         hash,
		Revision:     rec.Revision,
		PreviousHash: previousHash,
	}, nil
}
