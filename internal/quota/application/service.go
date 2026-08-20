package application

import (
	"context"
	"fmt"

	d "example.com/feature-rollout-control/internal/quota/domain"
)

type Store interface {
	Save(context.Context, d.Record) error
	Load(context.Context, string) (d.Record, error)
	UpdateQuotaReservation(context.Context, string, d.ReservationState, d.ReservationState) error
}

type Service struct{ store Store }

func New(s Store) *Service { return &Service{store: s} }

func (s *Service) Save(ctx context.Context, r d.Record) error {
	if !r.Valid() {
		return fmt.Errorf("invalid quota record")
	}
	return s.store.Save(ctx, r)
}

func (s *Service) Load(ctx context.Context, id string) (d.Record, error) {
	return s.store.Load(ctx, id)
}

func (s *Service) Reserve(ctx context.Context, id string) error {
	r, err := s.store.Load(ctx, id)
	if err != nil {
		return err
	}
	return s.store.UpdateQuotaReservation(ctx, id, r.State, d.Reserved)
}

func (s *Service) CompensateQuotaReservation(ctx context.Context, id string, commit func(context.Context) error) error {
	if err := commit(ctx); err != nil {
		// Publish failed: release the held quota so it can be reserved again
		// instead of leaving it occupied. The release is best-effort — the
		// original commit error is what we report to the caller.
		_ = s.store.UpdateQuotaReservation(ctx, id, d.Reserved, d.Released)
		return fmt.Errorf("commit quota: %w", err)
	}
	if err := s.store.UpdateQuotaReservation(ctx, id, d.Reserved, d.Committed); err != nil {
		return fmt.Errorf("commit reservation: %w", err)
	}
	return nil
}
