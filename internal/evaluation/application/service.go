package application

import (
	"context"
	"fmt"

	d "example.com/feature-rollout-control/internal/evaluation/domain"
)

type Store interface {
	Save(context.Context, d.Record) error
	Load(context.Context, string) (d.Record, error)
}

type Service struct{ store Store }

func New(s Store) *Service { return &Service{store: s} }

func (s *Service) Save(ctx context.Context, r d.Record) error {
	if !r.Valid() {
		return fmt.Errorf("invalid evaluation record")
	}
	if err := s.store.Save(ctx, r); err != nil {
		return fmt.Errorf("save evaluation: %w", err)
	}
	return nil
}

func (s *Service) Load(ctx context.Context, id string) (d.Record, error) {
	r, err := s.store.Load(ctx, id)
	if err != nil {
		return r, fmt.Errorf("load evaluation: %w", err)
	}
	return r, nil
}

func (s *Service) EvaluateBatchWithContext(ctx context.Context, ids []string, evaluate func(context.Context, string) error) error {
	for _, id := range ids {
		if err := evaluate(context.Background(), id); err != nil {
			return fmt.Errorf("evaluate %s: %w", id, err)
		}
	}
	return nil
}
