package infrastructure

import (
	"context"
	"fmt"
	"sync"

	d "example.com/feature-rollout-control/internal/distribution/domain"
)

type Memory struct {
	mu      sync.RWMutex
	records map[string]d.Record
	bundles map[string]d.Bundle
}

func New() *Memory {
	return &Memory{records: make(map[string]d.Record), bundles: make(map[string]d.Bundle)}
}

func (m *Memory) Save(ctx context.Context, r d.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[r.ID] = r
	return nil
}

func (m *Memory) Load(ctx context.Context, id string) (d.Record, error) {
	if err := ctx.Err(); err != nil {
		return d.Record{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.records[id]
	if !ok {
		return d.Record{}, fmt.Errorf("record %s not found", id)
	}
	return r, nil
}

func (m *Memory) SaveBundle(ctx context.Context, bundle d.Bundle) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bundles[bundle.Tenant] = d.CloneDistributionBundle(bundle)
	return nil
}

func (m *Memory) SnapshotDistribution(ctx context.Context, tenant string) (d.Bundle, error) {
	if err := ctx.Err(); err != nil {
		return d.Bundle{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	bundle, ok := m.bundles[tenant]
	if !ok {
		return d.Bundle{}, fmt.Errorf("tenant %s has no distribution bundle", tenant)
	}
	return d.CloneDistributionBundle(bundle), nil
}

func (m *Memory) UpdateEntry(ctx context.Context, tenant, key string, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	bundle, ok := m.bundles[tenant]
	if !ok {
		return fmt.Errorf("tenant %s has no distribution bundle", tenant)
	}
	bundle.Entries[key] = append([]byte(nil), value...)
	bundle.Version++
	m.bundles[tenant] = bundle
	return nil
}
