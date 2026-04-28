package channel

import (
	"context"
	"fmt"
	"os"

	"github.com/sroopra/ghega/pkg/channelstore"
)

// DiffResult compares a local channel definition with its deployed counterpart.
type DiffResult struct {
	ChannelName  string
	LocalHash    string
	DeployedHash string
	Identical    bool
	LocalYAML    string
	DeployedYAML string
}

// DiffLocal reads a local channel YAML file, hashes it, and compares it with
// the latest deployed revision in the store.
func DiffLocal(channelPath string, store channelstore.ChannelStore) (*DiffResult, error) {
	ctx := context.Background()

	data, err := os.ReadFile(channelPath)
	if err != nil {
		return nil, fmt.Errorf("read channel file: %w", err)
	}

	ch, valErrs := ValidateYAML(data)
	if len(valErrs) > 0 {
		return nil, fmt.Errorf("validation failed: %v", valErrs)
	}

	localHash, err := HashChannel(ch)
	if err != nil {
		return nil, fmt.Errorf("hash channel: %w", err)
	}

	result := &DiffResult{
		ChannelName: ch.Name,
		LocalHash:   localHash,
		LocalYAML:   string(data),
		Identical:   false,
	}

	deployed, err := store.GetChannel(ctx, ch.Name)
	if err != nil {
		// Not deployed yet.
		result.DeployedHash = ""
		return result, nil
	}

	result.DeployedHash = deployed.Hash
	result.DeployedYAML = string(deployed.YAML)
	result.Identical = localHash == deployed.Hash

	return result, nil
}
