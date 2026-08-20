package application

import (
	"context"
	"errors"
	"testing"

	d "example.com/feature-rollout-control/internal/project/domain"
	"example.com/feature-rollout-control/internal/project/infrastructure"
)

func TestProjectServicePreservesConflict(t *testing.T) {
	s := New(infrastructure.New())
	if err := s.RegisterProject(context.Background(), d.Project{ID: "p-1", TenantID: "acme", Name: "checkout"}); err != nil {
		t.Fatal(err)
	}
	err := s.RegisterProject(context.Background(), d.Project{ID: "p-2", TenantID: "acme", Name: "checkout"})
	if !errors.Is(err, d.ErrProjectConflict) {
		t.Fatalf("expected conflict identity, got %v", err)
	}
}
