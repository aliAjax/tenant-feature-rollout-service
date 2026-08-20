package infrastructure

import (
	"context"
	"errors"
	"testing"

	d "example.com/feature-rollout-control/internal/quota/domain"
)

func TestQuotaStoreRejectsStaleTransition(t *testing.T) {
	m := New()
	if err := m.Save(context.Background(), d.Record{ID: "q-1", Name: "production", State: d.Reserved}); err != nil {
		t.Fatal(err)
	}
	err := m.UpdateQuotaReservation(context.Background(), "q-1", d.Available, d.Reserved)
	if !errors.Is(err, ErrStateConflict) {
		t.Fatalf("expected stale transition conflict, got %v", err)
	}
	r, _ := m.Load(context.Background(), "q-1")
	if r.State != d.Reserved || r.Version != 0 {
		t.Fatalf("stale update changed record: %#v", r)
	}
}
