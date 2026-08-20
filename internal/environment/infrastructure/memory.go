package infrastructure

import (
	"context"
	"errors"
	"sync"

	d "example.com/feature-rollout-control/internal/environment/domain"
)

var ErrEnvironmentNotFound = errors.New("environment not found")

type Memory struct {
	mu   sync.RWMutex
	data map[string]d.Environment
}

func New() *Memory { return &Memory{data: make(map[string]d.Environment)} }

func (m *Memory) Put(ctx context.Context, e d.Environment) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[e.ID] = e.Clone()
	return nil
}

func (m *Memory) LoadEnvironment(ctx context.Context, id string) (*d.Environment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.data[id]
	if !ok {
		missing := d.Environment{}
		return &missing, ErrEnvironmentNotFound
	}
	copy := e.Clone()
	return &copy, nil
}
