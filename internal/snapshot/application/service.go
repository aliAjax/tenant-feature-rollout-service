package application

import (
	"context"
	"fmt"
	"io"

	d "example.com/feature-rollout-control/internal/snapshot/domain"
)

type Store interface {
	Save(context.Context, d.Record) error
	Load(context.Context, string) (d.Record, error)
}

type ResourceFactory interface {
	Open(context.Context, string) (io.WriteCloser, error)
}

type Service struct{ store Store }

func New(s Store) *Service { return &Service{store: s} }

func (s *Service) Save(ctx context.Context, r d.Record) error {
	if !r.Valid() {
		return fmt.Errorf("invalid snapshot record")
	}
	return s.store.Save(ctx, r)
}

func (s *Service) Load(ctx context.Context, id string) (d.Record, error) {
	return s.store.Load(ctx, id)
}

func exportOne(ctx context.Context, factory ResourceFactory, id string, write func(io.Writer, string) error) (err error) {
	resource, err := factory.Open(ctx, id)
	if err != nil {
		return err
	}
	return write(resource, id)
}

func (s *Service) ExportSnapshots(ctx context.Context, ids []string, factory ResourceFactory, write func(io.Writer, string) error) error {
	for _, id := range ids {
		if err := exportOne(ctx, factory, id, write); err != nil {
			return fmt.Errorf("export snapshot %s: %w", id, err)
		}
	}
	return nil
}
