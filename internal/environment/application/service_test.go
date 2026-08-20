package application

import (
	"context"
	"errors"
	"testing"

	d "example.com/feature-rollout-control/internal/environment/domain"
	"example.com/feature-rollout-control/internal/environment/infrastructure"
)

type optionalValidator struct{}

func (v *optionalValidator) Validate(*d.Environment) error {
	if v == nil {
		return nil
	}
	return nil
}

func TestEnvironmentServiceRejectsNilValidator(t *testing.T) {
	m := infrastructure.New()
	if err := m.Put(context.Background(), d.Environment{ID: "prod", ProjectID: "p-1", Name: "production"}); err != nil {
		t.Fatal(err)
	}
	var validator *optionalValidator
	err := New(m).ApplyEnvironmentPolicy(context.Background(), "prod", "approval", "required", validator)
	if !errors.Is(err, ErrNilValidator) {
		t.Fatalf("expected nil validator error, got %v", err)
	}
}
