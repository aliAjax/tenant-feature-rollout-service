package infrastructure

import (
	"errors"
	"testing"

	d "example.com/feature-rollout-control/internal/project/domain"
)

func TestProjectStorePreservesConflict(t *testing.T) {
	m := New()
	if err := m.CreateProject(d.Project{ID: "p-1", TenantID: "acme", Name: "checkout"}); err != nil {
		t.Fatal(err)
	}
	err := m.CreateProject(d.Project{ID: "p-2", TenantID: "acme", Name: "checkout"})
	if !errors.Is(err, d.ErrProjectConflict) {
		t.Fatalf("expected conflict identity, got %v", err)
	}
}
