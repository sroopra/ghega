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
	IsNoOp       bool
}

// Deploy reads a channel YAML file, validates it, computes its hash, and persists
// it via the given store. If the hash already exists the deployment is a no-op.
func Deploy(channelPath string, store channelstore.ChannelStore) (*DeployResult, error) {
	ctx := context.Background()

	data, err := os.ReadFile(channelPath)
	if err != nil {
		return nil, fmt.Errorf("read channel file: %w", err)
	}

	ch, valErrs := ValidateYAML(data)
	if ch != nil {
		valErrs = append(valErrs, ValidatePolicies(ch)...)
	}
	if len(valErrs) > 0 {
		return nil, fmt.Errorf("validation failed: %v", valErrs)
	}

	hash, err := HashChannel(ch)
	if err != nil {
		return nil, fmt.Errorf("hash channel: %w", err)
	}

	// Check if this hash is already deployed.
	existing, err := store.GetChannelRevision(ctx, ch.Name, hash)
	if err == nil && existing != nil {
		return &DeployResult{
			Name:     ch.Name,
			Hash:     hash,
			Revision: existing.Revision,
			IsNoOp:   true,
		}, nil
	}

	// Get the previously deployed version (if any) before saving.
	prev, _ := store.GetChannel(ctx, ch.Name)
	prevHash := ""
	if prev != nil {
		prevHash = prev.Hash
	}

	if err := store.SaveChannel(ctx, ch.Name, hash, data, 0); err != nil {
		return nil, fmt.Errorf("save channel: %w", err)
	}

	if err := store.SaveDeploymentAudit(ctx, ch.Name, hash, "deploy"); err != nil {
		return nil, fmt.Errorf("save deployment audit: %w", err)
	}

	// Fetch the saved revision to return the assigned revision number.
	saved, err := store.GetChannelRevision(ctx, ch.Name, hash)
	if err != nil {
		return nil, fmt.Errorf("fetch saved revision: %w", err)
	}

	return &DeployResult{
		Name:         ch.Name,
		Hash:         hash,
		Revision:     saved.Revision,
		PreviousHash: prevHash,
	}, nil
}
