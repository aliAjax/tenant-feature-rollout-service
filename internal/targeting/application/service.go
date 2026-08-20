package application

import (
	"context"
	"fmt"

	d "example.com/feature-rollout-control/internal/targeting/domain"
)

type Store interface {
	Save(context.Context, d.Record) error
	Load(context.Context, string) (d.Record, error)
}

type Service struct{ store Store }

func New(s Store) *Service { return &Service{store: s} }

func (s *Service) Save(ctx context.Context, r d.Record) error {
	if !r.Valid() {
		return fmt.Errorf("invalid targeting plan")
	}
	return s.store.Save(ctx, r)
}

func (s *Service) Load(ctx context.Context, id string) (d.Record, error) {
	return s.store.Load(ctx, id)
}

func (s *Service) CompileTargetingPlan(input []d.Segment) []d.Segment {
	out := make([]d.Segment, 0, len(input))
	for _, segment := range input {
		if !segment.Enabled {
			continue
		}
		out = append(out, segment)
	}
	return d.CloneSegments(out)
}
