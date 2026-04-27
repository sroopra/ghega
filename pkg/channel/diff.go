package channel

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sroopra/ghega/pkg/channelstore"
)

// DiffResult compares a local channel definition with its deployed version.
type DiffResult struct {
	ChannelName  string
	LocalHash    string
	DeployedHash string
	Identical    bool
	LocalYAML    string
	DeployedYAML string
}

// DiffLocal reads a local channel YAML file, hashes it, and compares it to
// the latest deployed revision in the store.
func DiffLocal(channelPath string, store channelstore.ChannelStore) (*DiffResult, error) {
	data, err := os.ReadFile(channelPath)
	if err != nil {
		return nil, fmt.Errorf("read channel file: %w", err)
	}

	ch, verrs := ValidateYAML(data)
	if len(verrs) > 0 {
		msgs := make([]string, len(verrs))
		for i, v := range verrs {
			msgs[i] = fmt.Sprintf("%s: %s", v.Field, v.Message)
		}
		return nil, fmt.Errorf("validation failed: %s", strings.Join(msgs, "; "))
	}

	localHash, err := HashChannel(ch)
	if err != nil {
		return nil, fmt.Errorf("hash channel: %w", err)
	}

	ctx := context.Background()
	deployed, err := store.GetChannel(ctx, ch.Name)
	if err != nil {
		if _, ok := err.(*channelstore.ErrNotFound); ok {
			return &DiffResult{
				ChannelName:  ch.Name,
				LocalHash:    localHash,
				DeployedHash: "",
				Identical:    false,
				LocalYAML:    string(data),
				DeployedYAML: "",
			}, nil
		}
		return nil, fmt.Errorf("get deployed channel: %w", err)
	}

	return &DiffResult{
		ChannelName:  ch.Name,
		LocalHash:    localHash,
		DeployedHash: deployed.Hash,
		Identical:    localHash == deployed.Hash,
		LocalYAML:    string(data),
		DeployedYAML: string(deployed.YAML),
	}, nil
}
