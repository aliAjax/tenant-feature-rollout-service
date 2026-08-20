package infrastructure

import (
	"context"
	"fmt"
	"sync"

	d "example.com/feature-rollout-control/internal/rollout/domain"
)

type Worker func(context.Context) d.Result

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

func (m *Memory) RunRolloutWorkers(ctx context.Context, workers []Worker) []d.Result {
	results := make(chan d.Result, len(workers))
	var wg sync.WaitGroup
	wg.Add(len(workers))
	for _, worker := range workers {
		worker := worker
		go func() {
			defer wg.Done()
			results <- worker(ctx)
		}()
	}
	wg.Wait()
	close(results)
	out := make([]d.Result, 0, len(workers))
	for result := range results {
		out = append(out, result)
	}
	return out
}
