package channel

import (
	"context"
	"fmt"

	"github.com/sroopra/ghega/pkg/channelstore"
)

// Rollback rolls a channel back to the specified hash. If toHash is empty it
// uses the second-most-recent revision.
func Rollback(channelName, toHash string, store channelstore.ChannelStore) error {
	ctx := context.Background()

	if toHash == "" {
		revs, err := store.ListChannelRevisions(ctx, channelName)
		if err != nil {
			return fmt.Errorf("list revisions: %w", err)
		}
		if len(revs) < 2 {
			return fmt.Errorf("cannot rollback: channel %q has fewer than 2 revisions", channelName)
		}
		toHash = revs[1].Hash
	}

	if _, err := store.GetChannelRevision(ctx, channelName, toHash); err != nil {
		return fmt.Errorf("verify hash exists: %w", err)
	}

	if err := store.RollbackChannel(ctx, channelName, toHash); err != nil {
		return fmt.Errorf("rollback channel: %w", err)
	}

	if err := store.SaveDeploymentAudit(ctx, channelName, toHash, "rollback"); err != nil {
		return fmt.Errorf("save rollback audit: %w", err)
	}

	return nil
}
