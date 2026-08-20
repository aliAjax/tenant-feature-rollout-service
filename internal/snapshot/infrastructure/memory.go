package infrastructure

import (
	"context"
	"fmt"
	"sync"

	d "example.com/feature-rollout-control/internal/snapshot/domain"
)

type Transaction struct {
	mu          sync.Mutex
	committed   bool
	rolledBack  bool
	commitErr   error
	rollbackErr error
}

func NewTransaction(commitErr, rollbackErr error) *Transaction {
	return &Transaction{commitErr: commitErr, rollbackErr: rollbackErr}
}

func (t *Transaction) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.commitErr == nil {
		t.committed = true
	}
	return t.commitErr
}

func (t *Transaction) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rolledBack = true
	return t.rollbackErr
}

func (t *Transaction) State() (committed, rolledBack bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.committed, t.rolledBack
}

func CommitSnapshot(ctx context.Context, tx *Transaction, work func(context.Context) error) error {
	if err := work(ctx); err != nil {
		// Work failed before commit: release the lease so resources do not
		// accumulate across a batch while preserving the original error.
		return d.CloseSnapshotLease(err, tx.Rollback)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit snapshot: %w", err)
	}
	return nil
}

type Memory struct {
	mu   sync.RWMutex
	data map[string]d.Record
}

func New() *Memory { return &Memory{data: make(map[string]d.Record)} }
func (m *Memory) Save(_ context.Context, r d.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[r.ID] = r
	return nil
}
func (m *Memory) Load(_ context.Context, id string) (d.Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.data[id]
	if !ok {
		return r, fmt.Errorf("record %s not found", id)
	}
	return r, nil
}
