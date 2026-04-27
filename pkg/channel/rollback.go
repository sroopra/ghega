package channel

import (
	"context"
	"fmt"

	"github.com/sroopra/ghega/pkg/channelstore"
)

// Rollback reverts a channel to a specific revision hash. If toHash is empty,
// it rolls back to the previous revision (second-most-recent).
func Rollback(channelName string, toHash string, store channelstore.ChannelStore) error {
	ctx := context.Background()

	if toHash == "" {
		revs, err := store.ListChannelRevisions(ctx, channelName)
		if err != nil {
			return fmt.Errorf("list revisions: %w", err)
		}
		if len(revs) < 2 {
			return fmt.Errorf("no previous revision available for channel %q", channelName)
		}
		toHash = revs[1].Hash
	}

	// Verify hash exists in revision history.
	_, err := store.GetChannelRevision(ctx, channelName, toHash)
	if err != nil {
		return fmt.Errorf("verify hash exists: %w", err)
	}

	if err := store.RollbackChannel(ctx, channelName, toHash); err != nil {
		return fmt.Errorf("rollback channel: %w", err)
	}

	if err := store.SaveDeploymentAudit(ctx, channelName, toHash, "rollback"); err != nil {
		return fmt.Errorf("save deployment audit: %w", err)
	}

	return nil
}
