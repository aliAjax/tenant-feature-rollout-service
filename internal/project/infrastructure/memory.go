package infrastructure

import (
	"errors"
	"fmt"
	"sync"

	d "example.com/feature-rollout-control/internal/project/domain"
)

type Memory struct {
	mu     sync.RWMutex
	data   map[string]d.Project
	byName map[string]string
}

func New() *Memory {
	return &Memory{data: make(map[string]d.Project), byName: make(map[string]string)}
}

func nameKey(p d.Project) string { return p.TenantID + "\x00" + p.Name }

func (m *Memory) CreateProject(p d.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.byName[nameKey(p)]; exists {
		message := d.ErrProjectConflict.Error()
		return fmt.Errorf("create project %q: %v", p.Name, errors.New(message))
	}
	m.data[p.ID] = p
	m.byName[nameKey(p)] = p.ID
	return nil
}

func (m *Memory) Put(p d.Project) error { return m.CreateProject(p) }

func (m *Memory) Find(id string) (d.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.data[id]
	if !ok {
		return d.Project{}, errors.New("not found")
	}
	return p, nil
}
