package application

import (
	"context"
	"errors"
	"testing"

	d "example.com/feature-rollout-control/internal/quota/domain"
	"example.com/feature-rollout-control/internal/quota/infrastructure"
)

func TestQuotaServiceCompensatesFailure(t *testing.T) {
	m := infrastructure.New()
	if err := m.Save(context.Background(), d.Record{ID: "q-1", Name: "production", State: d.Reserved}); err != nil {
		t.Fatal(err)
	}
	s := New(m)
	commitErr := errors.New("publisher rejected release")
	if err := s.CompensateQuotaReservation(context.Background(), "q-1", func(context.Context) error { return commitErr }); !errors.Is(err, commitErr) {
		t.Fatalf("expected commit failure, got %v", err)
	}
	if err := s.Reserve(context.Background(), "q-1"); err != nil {
		t.Fatalf("released reservation could not be reused: %v", err)
	}
	r, _ := s.Load(context.Background(), "q-1")
	if r.State != d.Reserved {
		t.Fatalf("expected reserved after retry, got %s", r.State)
	}
}
