package channel

import (
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestRollback_ToSpecificHash(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("rollback-test")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	first, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("first Deploy failed: %v", err)
	}

	ch.Description = "updated"
	path = writeTestChannel(t, dir, "channel.yaml", ch)
	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("second Deploy failed: %v", err)
	}

	if err := Rollback("rollback-test", first.Hash, store); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	audits, err := store.ListDeploymentAudit(nil, "rollback-test")
	if err != nil {
		t.Fatalf("ListDeploymentAudit failed: %v", err)
	}
	if len(audits) != 3 {
		t.Fatalf("len(audits) = %d, want 3", len(audits))
	}
	found := false
	for _, a := range audits {
		if a.Action == "rollback" && a.Hash == first.Hash {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected rollback audit for hash %q", first.Hash)
	}
}

func TestRollback_AutoHash(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("rollback-test")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	first, err := Deploy(path, store)
	if err != nil {
		t.Fatalf("first Deploy failed: %v", err)
	}

	ch.Description = "updated"
	path = writeTestChannel(t, dir, "channel.yaml", ch)
	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("second Deploy failed: %v", err)
	}

	if err := Rollback("rollback-test", "", store); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify it rolled back to the first hash.
	audits, err := store.ListDeploymentAudit(nil, "rollback-test")
	if err != nil {
		t.Fatalf("ListDeploymentAudit failed: %v", err)
	}
	found := false
	for _, a := range audits {
		if a.Action == "rollback" && a.Hash == first.Hash {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected auto-rollback to hash %q", first.Hash)
	}
}

func TestRollback_AutoHash_NotEnoughRevisions(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("rollback-test")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	err := Rollback("rollback-test", "", store)
	if err == nil {
		t.Fatal("expected error when fewer than 2 revisions exist")
	}
}

func TestRollback_InvalidHash(t *testing.T) {
	store := channelstore.NewInMemoryStore()
	dir := t.TempDir()
	ch := testChannel("rollback-test")
	path := writeTestChannel(t, dir, "channel.yaml", ch)

	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	err := Rollback("rollback-test", "nonexistent-hash", store)
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
}
