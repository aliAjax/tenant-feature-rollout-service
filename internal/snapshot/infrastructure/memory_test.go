package infrastructure

import (
	"context"
	"errors"
	"testing"
)

func TestSnapshotTransactionRollsBack(t *testing.T) {
	primary := errors.New("snapshot write rejected")
	tx := NewTransaction(errors.New("commit unavailable"), nil)
	err := CommitSnapshot(context.Background(), tx, func(context.Context) error { return primary })
	committed, rolledBack := tx.State()
	if !errors.Is(err, primary) || committed || !rolledBack {
		t.Fatalf("unexpected transaction result err=%v committed=%v rolledBack=%v", err, committed, rolledBack)
	}
}
