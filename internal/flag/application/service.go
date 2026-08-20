package application

import (
	"context"
	"fmt"

	d "example.com/feature-rollout-control/internal/flag/domain"
)

type Store interface {
	Save(context.Context, d.Record) error
	Load(context.Context, string) (d.Record, error)
	CompareAndSwapFlagState(context.Context, string, d.State, d.State) error
}

type Service struct{ store Store }

func New(s Store) *Service { return &Service{store: s} }

func (s *Service) Save(ctx context.Context, r d.Record) error {
	if !r.Valid() {
		return fmt.Errorf("invalid flag")
	}
	return s.store.Save(ctx, r)
}

func (s *Service) Load(ctx context.Context, id string) (d.Record, error) {
	return s.store.Load(ctx, id)
}

func (s *Service) PublishFlag(ctx context.Context, id string) (d.Record, error) {
	r, err := s.store.Load(ctx, id)
	if err != nil {
		return d.Record{}, err
	}
	if d.CanTransitionFlag(r.State, d.Published) == false {
		return d.Record{}, fmt.Errorf("publish %s from %s: invalid transition", id, r.State)
	}
	if err := s.store.CompareAndSwapFlagState(ctx, id, r.State, d.Published); err != nil {
		return d.Record{}, fmt.Errorf("publish %s: %w", id, err)
	}
	return r, nil
}
