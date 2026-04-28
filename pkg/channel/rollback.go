package channel

import (
	"context"
	"fmt"

	"github.com/sroopra/ghega/pkg/channelstore"
)

// Rollback reverts a channel to a specific revision hash in the store by
// re-saving that revision as the latest.
func Rollback(ctx context.Context, store channelstore.ChannelStore, name, hash string) error {
	rec, err := store.GetChannelRevision(ctx, name, hash)
	if err != nil {
		return fmt.Errorf("get channel revision: %w", err)
	}
	if err := store.SaveChannel(ctx, name, hash, rec.YAML, 0); err != nil {
		return fmt.Errorf("save channel: %w", err)
	}
	return store.RollbackChannel(ctx, name, hash)
}
