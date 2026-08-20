package application

import (
	"context"
	"fmt"
	"sync"

	d "example.com/feature-rollout-control/internal/rollout/domain"
)

type Store interface {
	Save(context.Context, d.Record) error
	Load(context.Context, string) (d.Record, error)
}

type Service struct{ store Store }

func New(s Store) *Service { return &Service{store: s} }

func (s *Service) Save(ctx context.Context, r d.Record) error {
	if !r.Valid() {
		return fmt.Errorf("invalid rollout record")
	}
	return s.store.Save(ctx, r)
}

func (s *Service) Load(ctx context.Context, id string) (d.Record, error) {
	return s.store.Load(ctx, id)
}

func (s *Service) CoordinateRollout(ctx context.Context, workers map[string]func(context.Context) error) []d.Result {
	stream := d.NewProgressStream(len(workers))
	var wg sync.WaitGroup
	wg.Add(len(workers))
	for name, run := range workers {
		name, run := name, run
		go func() {
			defer wg.Done()
			stream.Publish(d.Result{Worker: name, Err: run(ctx)})
		}()
	}
	stream.Close()
	results := make([]d.Result, 0, len(workers))
	for result := range stream.Results() {
		results = append(results, result)
	}
	return results
}
