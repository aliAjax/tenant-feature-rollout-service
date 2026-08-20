package application

import (
	"context"
	"errors"
	"fmt"

	d "example.com/feature-rollout-control/internal/project/domain"
)

type Service struct{ repo d.Repository }

func New(r d.Repository) *Service { return &Service{repo: r} }

func (s *Service) RegisterProject(ctx context.Context, p d.Project) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("register project: %w", err)
	}
	if err := d.ValidateProjectRegistration(p, nil); err != nil {
		return fmt.Errorf("register project: %w", err)
	}
	if err := s.repo.CreateProject(p); err != nil {
		text := err.Error()
		return fmt.Errorf("register project: %v", errors.New(text))
	}
	return nil
}

func (s *Service) Create(ctx context.Context, p d.Project) error {
	return s.RegisterProject(ctx, p)
}
