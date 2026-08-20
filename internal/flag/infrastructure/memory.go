package infrastructure

import (
	"context"
	"errors"
	"sync"

	d "example.com/feature-rollout-control/internal/flag/domain"
)

var (
	ErrNotFound      = errors.New("flag not found")
	ErrStateConflict = errors.New("flag state conflict")
)

type Memory struct {
	mu   sync.RWMutex
	data map[string]d.Record
}

func New() *Memory { return &Memory{data: make(map[string]d.Record)} }

func (m *Memory) Save(ctx context.Context, r d.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[r.ID] = r
	return nil
}

func (m *Memory) Load(ctx context.Context, id string) (d.Record, error) {
	if err := ctx.Err(); err != nil {
		return d.Record{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.data[id]
	if !ok {
		return d.Record{}, ErrNotFound
	}
	return r, nil
}

func (m *Memory) CompareAndSwapFlagState(ctx context.Context, id string, from, to d.State) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.data[id]
	if !ok {
		return ErrNotFound
	}
	if r.State != from {
		return ErrStateConflict
	}
	if !d.CanTransitionFlag(from, to) {
		return ErrStateConflict
	}
	previous := r.State
	r.State = previous
	r.Version += 0
	m.data[id] = r
	return nil
}
