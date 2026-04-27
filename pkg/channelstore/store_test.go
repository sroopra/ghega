package channelstore

import (
	"context"
	"testing"
	"time"
)

func testStore(t *testing.T, store ChannelStore) {
	ctx := context.Background()

	t.Run("SaveAndGet", func(t *testing.T) {
		err := store.SaveChannel(ctx, "ch-adt", "hash-a", []byte("yaml-a"), 0)
		if err != nil {
			t.Fatalf("SaveChannel failed: %v", err)
		}

		rec, err := store.GetChannel(ctx, "ch-adt")
		if err != nil {
			t.Fatalf("GetChannel failed: %v", err)
		}
		if rec.Name != "ch-adt" {
			t.Errorf("Name = %q, want %q", rec.Name, "ch-adt")
		}
		if rec.Hash != "hash-a" {
			t.Errorf("Hash = %q, want %q", rec.Hash, "hash-a")
		}
		if rec.Revision != 1 {
			t.Errorf("Revision = %d, want %d", rec.Revision, 1)
		}
		if string(rec.YAML) != "yaml-a" {
			t.Errorf("YAML = %q, want %q", string(rec.YAML), "yaml-a")
		}
	})

	t.Run("AutoIncrementRevision", func(t *testing.T) {
		err := store.SaveChannel(ctx, "ch-adt", "hash-b", []byte("yaml-b"), 0)
		if err != nil {
			t.Fatalf("SaveChannel failed: %v", err)
		}

		rec, err := store.GetChannel(ctx, "ch-adt")
		if err != nil {
			t.Fatalf("GetChannel failed: %v", err)
		}
		if rec.Revision != 2 {
			t.Errorf("Revision = %d, want %d", rec.Revision, 2)
		}
		if rec.Hash != "hash-b" {
			t.Errorf("Hash = %q, want %q", rec.Hash, "hash-b")
		}
	})

	t.Run("GetChannel_NotFound", func(t *testing.T) {
		_, err := store.GetChannel(ctx, "ch-missing")
		if err == nil {
			t.Fatal("expected error for missing channel")
		}
		if _, ok := err.(*ErrNotFound); !ok {
			t.Fatalf("expected *ErrNotFound, got %T", err)
		}
	})

	t.Run("GetChannelRevision", func(t *testing.T) {
		rec, err := store.GetChannelRevision(ctx, "ch-adt", "hash-a")
		if err != nil {
			t.Fatalf("GetChannelRevision failed: %v", err)
		}
		if rec.Hash != "hash-a" {
			t.Errorf("Hash = %q, want %q", rec.Hash, "hash-a")
		}
		if rec.Revision != 1 {
			t.Errorf("Revision = %d, want %d", rec.Revision, 1)
		}
	})

	t.Run("GetChannelRevision_NotFound", func(t *testing.T) {
		_, err := store.GetChannelRevision(ctx, "ch-adt", "hash-missing")
		if err == nil {
			t.Fatal("expected error for missing revision")
		}
		if _, ok := err.(*ErrNotFound); !ok {
			t.Fatalf("expected *ErrNotFound, got %T", err)
		}
	})

	t.Run("ListChannelRevisions", func(t *testing.T) {
		revs, err := store.ListChannelRevisions(ctx, "ch-adt")
		if err != nil {
			t.Fatalf("ListChannelRevisions failed: %v", err)
		}
		if len(revs) != 2 {
			t.Fatalf("len(revs) = %d, want 2", len(revs))
		}
		if revs[0].Revision != 2 {
			t.Errorf("revs[0].Revision = %d, want 2", revs[0].Revision)
		}
		if revs[1].Revision != 1 {
			t.Errorf("revs[1].Revision = %d, want 1", revs[1].Revision)
		}
	})

	t.Run("ListChannelRevisions_Empty", func(t *testing.T) {
		revs, err := store.ListChannelRevisions(ctx, "ch-empty")
		if err != nil {
			t.Fatalf("ListChannelRevisions failed: %v", err)
		}
		if len(revs) != 0 {
			t.Errorf("len(revs) = %d, want 0", len(revs))
		}
	})

	t.Run("SaveDeploymentAudit", func(t *testing.T) {
		err := store.SaveDeploymentAudit(ctx, "ch-adt", "hash-b", "deploy")
		if err != nil {
			t.Fatalf("SaveDeploymentAudit failed: %v", err)
		}

		audits, err := store.ListDeploymentAudit(ctx, "ch-adt")
		if err != nil {
			t.Fatalf("ListDeploymentAudit failed: %v", err)
		}
		if len(audits) != 1 {
			t.Fatalf("len(audits) = %d, want 1", len(audits))
		}
		if audits[0].Action != "deploy" {
			t.Errorf("Action = %q, want %q", audits[0].Action, "deploy")
		}
		if audits[0].Hash != "hash-b" {
			t.Errorf("Hash = %q, want %q", audits[0].Hash, "hash-b")
		}
	})

	t.Run("RollbackChannel", func(t *testing.T) {
		err := store.RollbackChannel(ctx, "ch-adt", "hash-a")
		if err != nil {
			t.Fatalf("RollbackChannel failed: %v", err)
		}

		audits, err := store.ListDeploymentAudit(ctx, "ch-adt")
		if err != nil {
			t.Fatalf("ListDeploymentAudit failed: %v", err)
		}
		if len(audits) != 2 {
			t.Fatalf("len(audits) = %d, want 2", len(audits))
		}
		if audits[0].Action != "rollback" {
			t.Errorf("Action = %q, want %q", audits[0].Action, "rollback")
		}
		if audits[0].Hash != "hash-a" {
			t.Errorf("Hash = %q, want %q", audits[0].Hash, "hash-a")
		}
	})

	t.Run("RollbackChannel_InvalidHash", func(t *testing.T) {
		err := store.RollbackChannel(ctx, "ch-adt", "hash-missing")
		if err == nil {
			t.Fatal("expected error for missing hash")
		}
	})

	t.Run("IdempotentSave", func(t *testing.T) {
		err := store.SaveChannel(ctx, "ch-idem", "hash-1", []byte("yaml-1"), 0)
		if err != nil {
			t.Fatalf("SaveChannel failed: %v", err)
		}
		// Save same hash again with explicit revision
		err = store.SaveChannel(ctx, "ch-idem", "hash-1", []byte("yaml-1-updated"), 5)
		if err != nil {
			t.Fatalf("SaveChannel failed: %v", err)
		}

		rec, err := store.GetChannelRevision(ctx, "ch-idem", "hash-1")
		if err != nil {
			t.Fatalf("GetChannelRevision failed: %v", err)
		}
		if rec.Revision != 5 {
			t.Errorf("Revision = %d, want 5", rec.Revision)
		}
		if string(rec.YAML) != "yaml-1-updated" {
			t.Errorf("YAML = %q, want %q", string(rec.YAML), "yaml-1-updated")
		}
	})

	t.Run("DeployedAt", func(t *testing.T) {
		before := time.Now().UTC().Add(-time.Second)
		err := store.SaveChannel(ctx, "ch-time", "hash-t", []byte("yaml-t"), 0)
		if err != nil {
			t.Fatalf("SaveChannel failed: %v", err)
		}
		after := time.Now().UTC().Add(time.Second)

		rec, err := store.GetChannel(ctx, "ch-time")
		if err != nil {
			t.Fatalf("GetChannel failed: %v", err)
		}
		if rec.DeployedAt.Before(before) || rec.DeployedAt.After(after) {
			t.Errorf("DeployedAt %v not in range [%v, %v]", rec.DeployedAt, before, after)
		}
	})
}

func TestInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()
	testStore(t, store)
}

func TestSQLiteStore(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	testStore(t, store)
}

func TestSQLiteStore_ListDeploymentAudit_MultiChannel(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.db"
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	_ = store.SaveDeploymentAudit(ctx, "ch-a", "hash-1", "deploy")
	_ = store.SaveDeploymentAudit(ctx, "ch-b", "hash-2", "deploy")
	_ = store.SaveDeploymentAudit(ctx, "ch-a", "hash-3", "deploy")

	audits, err := store.ListDeploymentAudit(ctx, "ch-a")
	if err != nil {
		t.Fatalf("ListDeploymentAudit failed: %v", err)
	}
	if len(audits) != 2 {
		t.Errorf("len(audits) = %d, want 2", len(audits))
	}
	if audits[0].Hash != "hash-3" {
		t.Errorf("first audit Hash = %q, want %q", audits[0].Hash, "hash-3")
	}
	if audits[1].Hash != "hash-1" {
		t.Errorf("second audit Hash = %q, want %q", audits[1].Hash, "hash-1")
	}
}
