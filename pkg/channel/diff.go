package channel

import (
	"context"
	"fmt"
	"os"

	"github.com/sroopra/ghega/pkg/channelstore"
)

// DiffResult holds the outcome of comparing a local channel definition with
// the currently deployed revision.
type DiffResult struct {
	ChannelName  string
	LocalHash    string
	DeployedHash string
	Identical    bool
	LocalYAML    string
	DeployedYAML string
}

// DiffLocal reads a local channel YAML file, validates it, computes its hash,
// and compares it with the currently deployed revision from the store.
func DiffLocal(channelPath string, store channelstore.ChannelStore) (*DiffResult, error) {
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

	localHash, err := HashChannel(ch)
	if err != nil {
		return nil, fmt.Errorf("hash channel: %w", err)
	}

	ctx := context.Background()

	result := &DiffResult{
		ChannelName: ch.Name,
		LocalHash:   localHash,
		LocalYAML:   string(data),
		Identical:   false,
	}

	current, err := store.GetChannel(ctx, ch.Name)
	if err != nil {
		if _, ok := err.(*channelstore.ErrNotFound); ok {
			return result, nil
		}
		return nil, fmt.Errorf("get deployed channel: %w", err)
	}

	result.DeployedHash = current.Hash
	result.DeployedYAML = string(current.YAML)
	result.Identical = localHash == current.Hash

	return result, nil
}
