package infrastructure

import (
	"context"
	"errors"
	"sync"

	d "example.com/feature-rollout-control/internal/targeting/domain"
)

var ErrNotFound = errors.New("targeting plan not found")

type Memory struct {
	mu   sync.RWMutex
	data map[string]d.Record
}

func New() *Memory { return &Memory{data: make(map[string]d.Record)} }

func (m *Memory) SaveTargetingPlan(ctx context.Context, r d.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[r.ID] = r.Clone()
	return nil
}

func (m *Memory) Save(ctx context.Context, r d.Record) error { return m.SaveTargetingPlan(ctx, r) }

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
	return r.Clone(), nil
}
