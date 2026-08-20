package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"time"

	d "example.com/feature-rollout-control/internal/evaluation/domain"
)

type Memory struct {
	mu   sync.RWMutex
	data map[string]d.Record
	gate <-chan struct{}
}

func New() *Memory {
	ready := make(chan struct{})
	close(ready)
	return &Memory{data: make(map[string]d.Record), gate: ready}
}

func NewWithGate(gate <-chan struct{}) *Memory {
	return &Memory{data: make(map[string]d.Record), gate: gate}
}

func (m *Memory) Save(ctx context.Context, r d.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[r.ID] = r
	return nil
}

func (m *Memory) LoadEvaluationContext(ctx context.Context, id string) (d.Record, error) {
	select {
	case <-m.gate:
	case <-time.After(200 * time.Millisecond):
		return d.Record{}, fmt.Errorf("evaluation store remained blocked")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.data[id]
	if !ok {
		return d.Record{}, fmt.Errorf("record %s not found", id)
	}
	return r, nil
}

func (m *Memory) Load(ctx context.Context, id string) (d.Record, error) {
	return m.LoadEvaluationContext(ctx, id)
}
