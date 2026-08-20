package infrastructure

import (
	"context"
	"testing"

	d "example.com/feature-rollout-control/internal/flag/domain"
)

func TestFlagStoreCompareAndSwap(t *testing.T) {
	m := New()
	if err := m.Save(context.Background(), d.Record{ID: "flag-1", Name: "checkout", State: d.Review}); err != nil {
		t.Fatal(err)
	}
	if err := m.CompareAndSwapFlagState(context.Background(), "flag-1", d.Review, d.Published); err != nil {
		t.Fatal(err)
	}
	r, _ := m.Load(context.Background(), "flag-1")
	if r.State != d.Published || r.Version != 1 {
		t.Fatalf("unexpected record: %#v", r)
	}
}
