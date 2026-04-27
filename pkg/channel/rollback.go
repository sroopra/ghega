package channel

import (
	"context"
	"fmt"

	"github.com/sroopra/ghega/pkg/channelstore"
)

// Rollback rolls a channel back to a specific revision hash. If toHash is
// empty, it rolls back to the previous revision (second-most-recent).
func Rollback(channelName string, toHash string, store channelstore.ChannelStore) error {
	ctx := context.Background()

	if toHash == "" {
		revs, err := store.ListChannelRevisions(ctx, channelName)
		if err != nil {
			return fmt.Errorf("list revisions: %w", err)
		}
		if len(revs) < 2 {
			return fmt.Errorf("no previous revision to roll back to")
		}
		toHash = revs[1].Hash
	}

	if _, err := store.GetChannelRevision(ctx, channelName, toHash); err != nil {
		return fmt.Errorf("verify hash exists: %w", err)
	}

	if err := store.RollbackChannel(ctx, channelName, toHash); err != nil {
		return fmt.Errorf("rollback channel: %w", err)
	}

	return nil
}
