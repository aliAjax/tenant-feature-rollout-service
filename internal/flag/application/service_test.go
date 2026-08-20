package application

import (
	"context"
	"testing"

	d "example.com/feature-rollout-control/internal/flag/domain"
	"example.com/feature-rollout-control/internal/flag/infrastructure"
)

func TestFlagServicePublishesReviewed(t *testing.T) {
	m := infrastructure.New()
	if err := m.Save(context.Background(), d.Record{ID: "flag-1", Name: "checkout", State: d.Review}); err != nil {
		t.Fatal(err)
	}
	r, err := New(m).PublishFlag(context.Background(), "flag-1")
	if err != nil {
		t.Fatal(err)
	}
	if r.State != d.Published {
		t.Fatalf("expected published, got %s", r.State)
	}
}
