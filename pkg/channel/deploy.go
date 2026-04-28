package channel

import (
	"context"
	"fmt"

	"github.com/sroopra/ghega/pkg/channelstore"
)

// DeployResult holds the outcome of a channel deployment.
type DeployResult struct {
	Name     string
	Revision int
	Hash     string
	NoOp     bool
}

// Deploy persists a channel to the store. If the channel hash already matches
// the latest deployed revision, it returns a no-op result.
func Deploy(ctx context.Context, store channelstore.ChannelStore, ch *Channel, yamlBytes []byte) (*DeployResult, error) {
	hash, err := HashChannel(ch)
	if err != nil {
		return nil, fmt.Errorf("hash channel: %w", err)
	}

	existing, err := store.GetChannel(ctx, ch.Name)
	if err == nil && existing.Hash == hash {
		return &DeployResult{
			Name:     ch.Name,
			Revision: existing.Revision,
			Hash:     hash,
			NoOp:     true,
		}, nil
	}

	if err := store.SaveChannel(ctx, ch.Name, hash, yamlBytes, 0); err != nil {
		return nil, fmt.Errorf("save channel: %w", err)
	}

	saved, err := store.GetChannel(ctx, ch.Name)
	if err != nil {
		return nil, fmt.Errorf("get saved channel: %w", err)
	}

	return &DeployResult{
		Name:     ch.Name,
		Revision: saved.Revision,
		Hash:     hash,
		NoOp:     false,
	}, nil
}
