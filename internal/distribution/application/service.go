package application

import (
	"context"
	"fmt"
	"sync"

	d "example.com/feature-rollout-control/internal/distribution/domain"
)

type Store interface {
	Save(context.Context, d.Record) error
	Load(context.Context, string) (d.Record, error)
	SnapshotDistribution(context.Context, string) (d.Bundle, error)
}

type Service struct{ store Store }

func New(s Store) *Service { return &Service{store: s} }

func (s *Service) Save(ctx context.Context, r d.Record) error {
	if !r.Valid() {
		return fmt.Errorf("invalid distribution record")
	}
	return s.store.Save(ctx, r)
}

func (s *Service) Load(ctx context.Context, id string) (d.Record, error) {
	return s.store.Load(ctx, id)
}

func (s *Service) FanOutDistribution(ctx context.Context, tenant string, subscribers []func(d.Bundle)) error {
	snapshot, err := s.store.SnapshotDistribution(ctx, tenant)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(len(subscribers))
	for _, subscriber := range subscribers {
		subscriber := subscriber
		bundle := snapshot
		go func() {
			defer wg.Done()
			subscriber(bundle)
		}()
	}
	wg.Wait()
	return nil
}
