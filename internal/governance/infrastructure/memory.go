package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"sync"

	d "example.com/feature-rollout-control/internal/governance/domain"
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
		return d.Record{}, fmt.Errorf("record %s not found", id)
	}
	return r, nil
}

func (m *Memory) AppendGovernanceAudit(ctx context.Context, event string, write func(context.Context, string) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := write(ctx, event); err != nil {
		message := err.Error()
		return fmt.Errorf("append governance audit %q: %v", event, errors.New(message))
	}
	return nil
}
