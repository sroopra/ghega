package channelstore

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// InMemoryStore is a development/testing implementation of ChannelStore.
type InMemoryStore struct {
	mu         sync.RWMutex
	channels   map[string][]ChannelRecord // keyed by name, sorted by revision asc
	deployments []AuditRecord
}

// NewInMemoryStore creates a new empty InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		channels:    make(map[string][]ChannelRecord),
		deployments: make([]AuditRecord, 0),
	}
}

// SaveChannel persists a channel revision. If revision <= 0 it is auto-incremented.
func (s *InMemoryStore) SaveChannel(_ context.Context, name, hash string, yamlBytes []byte, revision int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	revs := s.channels[name]
	if revision <= 0 {
		maxRev := 0
		for _, r := range revs {
			if r.Revision > maxRev {
				maxRev = r.Revision
			}
		}
		revision = maxRev + 1
	}

	// Upsert: replace existing record with same hash, otherwise append.
	replaced := false
	for i := range revs {
		if revs[i].Hash == hash {
			revs[i] = ChannelRecord{
				Name:       name,
				Hash:       hash,
				YAML:       append([]byte(nil), yamlBytes...),
				Revision:   revision,
				DeployedAt: time.Now().UTC(),
			}
			replaced = true
			break
		}
	}
	if !replaced {
		revs = append(revs, ChannelRecord{
			Name:       name,
			Hash:       hash,
			YAML:       append([]byte(nil), yamlBytes...),
			Revision:   revision,
			DeployedAt: time.Now().UTC(),
		})
	}
	s.channels[name] = revs
	return nil
}

// GetChannel returns the latest revision for a channel name.
func (s *InMemoryStore) GetChannel(_ context.Context, name string) (*ChannelRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	revs, ok := s.channels[name]
	if !ok || len(revs) == 0 {
		return nil, &ErrNotFound{Name: name}
	}

	latest := revs[0]
	for _, r := range revs {
		if r.Revision > latest.Revision {
			latest = r
		}
	}
	cp := latest
	cp.YAML = append([]byte(nil), cp.YAML...)
	return &cp, nil
}

// GetChannelRevision returns a specific revision by hash.
func (s *InMemoryStore) GetChannelRevision(_ context.Context, name, hash string) (*ChannelRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	revs, ok := s.channels[name]
	if !ok {
		return nil, &ErrNotFound{Name: name, Hash: hash}
	}
	for _, r := range revs {
		if r.Hash == hash {
			cp := r
			cp.YAML = append([]byte(nil), cp.YAML...)
			return &cp, nil
		}
	}
	return nil, &ErrNotFound{Name: name, Hash: hash}
}

// ListChannelRevisions returns all revisions for a channel ordered by revision desc.
func (s *InMemoryStore) ListChannelRevisions(_ context.Context, name string) ([]ChannelRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	revs, ok := s.channels[name]
	if !ok {
		return nil, nil
	}

	out := make([]ChannelRecord, 0, len(revs))
	for _, r := range revs {
		cp := r
		cp.YAML = append([]byte(nil), cp.YAML...)
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Revision > out[j].Revision
	})
	return out, nil
}

// RollbackChannel verifies the hash exists, re-saves it as the latest revision,
// and records a rollback audit entry.
func (s *InMemoryStore) RollbackChannel(ctx context.Context, name, hash string) error {
	rec, err := s.GetChannelRevision(ctx, name, hash)
	if err != nil {
		return err
	}
	// Re-save with a new auto-incremented revision so it becomes the latest.
	if err := s.SaveChannel(ctx, name, hash, rec.YAML, 0); err != nil {
		return fmt.Errorf("re-save rolled-back revision: %w", err)
	}
	return s.SaveDeploymentAudit(ctx, name, hash, "rollback")
}

// SaveDeploymentAudit records a deployment action.
func (s *InMemoryStore) SaveDeploymentAudit(_ context.Context, channelName, hash, action string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deployments = append(s.deployments, AuditRecord{
		ChannelName: channelName,
		Hash:        hash,
		Action:      action,
		Timestamp:   time.Now().UTC(),
	})
	return nil
}

// ListDeploymentAudit returns audit entries for a channel ordered by timestamp desc.
func (s *InMemoryStore) ListDeploymentAudit(_ context.Context, channelName string) ([]AuditRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []AuditRecord
	for i := len(s.deployments) - 1; i >= 0; i-- {
		if s.deployments[i].ChannelName == channelName {
			out = append(out, s.deployments[i])
		}
	}
	return out, nil
}
