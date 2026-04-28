package channel

import (
	"context"
	"fmt"

	"github.com/sroopra/ghega/pkg/channelstore"
)

// DiffResult holds the outcome of comparing a local channel with its deployed
// revision.
type DiffResult struct {
	Name         string
	LocalHash    string
	DeployedHash string
	Identical    bool
	Deployed     bool
}

// DiffLocal compares the local channel definition with the latest deployed
// revision in the store.
func DiffLocal(ctx context.Context, store channelstore.ChannelStore, ch *Channel) (*DiffResult, error) {
	localHash, err := HashChannel(ch)
	if err != nil {
		return nil, fmt.Errorf("hash local channel: %w", err)
	}

	deployed, err := store.GetChannel(ctx, ch.Name)
	if err != nil {
		if _, ok := err.(*channelstore.ErrNotFound); ok {
			return &DiffResult{
				Name:      ch.Name,
				LocalHash: localHash,
				Identical: false,
				Deployed:  false,
			}, nil
		}
		return nil, fmt.Errorf("get deployed channel: %w", err)
	}

	return &DiffResult{
		Name:         ch.Name,
		LocalHash:    localHash,
		DeployedHash: deployed.Hash,
		Identical:    localHash == deployed.Hash,
		Deployed:     true,
	}, nil
}
