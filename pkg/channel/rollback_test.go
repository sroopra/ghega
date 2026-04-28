package channel

import (
	"testing"

	"github.com/sroopra/ghega/pkg/channelstore"
)

func TestRollback_ByHash(t *testing.T) {
	dir := t.TempDir()
	store := channelstore.NewInMemoryStore()
	path1 := makeTempChannel(t, dir, "roll-ch", "")

	res1, err := Deploy(path1, store)
	if err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	path2 := makeTempChannel(t, dir, "roll-ch", "  - source: MSH-9.1\n    target: message_type\n")
	if _, err := Deploy(path2, store); err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	if err := Rollback("roll-ch", res1.Hash, store); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	audits, err := store.ListDeploymentAudit(nil, "roll-ch")
	if err != nil {
		t.Fatalf("list audits: %v", err)
	}
	found := false
	for _, a := range audits {
		if a.Action == "rollback" && a.Hash == res1.Hash {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected rollback audit entry")
	}
}

func TestRollback_Auto(t *testing.T) {
	dir := t.TempDir()
	store := channelstore.NewInMemoryStore()
	path1 := makeTempChannel(t, dir, "roll-ch", "")

	res1, err := Deploy(path1, store)
	if err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	path2 := makeTempChannel(t, dir, "roll-ch", "  - source: MSH-9.1\n    target: message_type\n")
	if _, err := Deploy(path2, store); err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	if err := Rollback("roll-ch", "", store); err != nil {
		t.Fatalf("rollback auto: %v", err)
	}

	audits, err := store.ListDeploymentAudit(nil, "roll-ch")
	if err != nil {
		t.Fatalf("list audits: %v", err)
	}
	found := false
	for _, a := range audits {
		if a.Action == "rollback" && a.Hash == res1.Hash {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected rollback audit entry for previous hash")
	}
}

func TestRollback_AutoTooFew(t *testing.T) {
	dir := t.TempDir()
	store := channelstore.NewInMemoryStore()
	path := makeTempChannel(t, dir, "roll-ch", "")
	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	err := Rollback("roll-ch", "", store)
	if err == nil {
		t.Fatal("expected error when fewer than 2 revisions")
	}
}

func TestRollback_InvalidHash(t *testing.T) {
	dir := t.TempDir()
	store := channelstore.NewInMemoryStore()
	path := makeTempChannel(t, dir, "roll-ch", "")
	if _, err := Deploy(path, store); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	err := Rollback("roll-ch", "0000000000000000000000000000000000000000000000000000000000000000", store)
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
}
