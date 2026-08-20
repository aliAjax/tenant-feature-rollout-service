package domain

import (
	"errors"
	"testing"
)

func TestProjectValidationPreservesConflict(t *testing.T) {
	err := ValidateProjectRegistration(Project{ID: "p-2", TenantID: "acme", Name: "checkout"}, ErrProjectConflict)
	if !errors.Is(err, ErrProjectConflict) {
		t.Fatalf("expected conflict identity, got %v", err)
	}
}
