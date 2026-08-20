package application

import (
	"context"
	"errors"
	"fmt"

	d "example.com/feature-rollout-control/internal/governance/domain"
)

type Store interface {
	Save(context.Context, d.Record) error
	Load(context.Context, string) (d.Record, error)
}

type Service struct{ store Store }

func New(s Store) *Service { return &Service{store: s} }

func (s *Service) Save(ctx context.Context, r d.Record) error {
	if !r.Valid() {
		return fmt.Errorf("invalid governance record")
	}
	return s.store.Save(ctx, r)
}

func (s *Service) Load(ctx context.Context, id string) (d.Record, error) {
	return s.store.Load(ctx, id)
}

func (s *Service) EvaluateGovernance(policy string, decisionErr, auditErr error) error {
	decision := d.WrapDecisionError(policy, decisionErr)
	if decision != nil && auditErr != nil {
		return errors.Join(decision, fmt.Errorf("governance audit warning: %w", auditErr))
	}
	if decision != nil {
		return decision
	}
	if auditErr != nil {
		return fmt.Errorf("governance audit warning: %w", auditErr)
	}
	return nil
}
