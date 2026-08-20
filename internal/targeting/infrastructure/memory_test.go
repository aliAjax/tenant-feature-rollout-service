package infrastructure

import (
	"context"
	"testing"

	d "example.com/feature-rollout-control/internal/targeting/domain"
)

func TestTargetingStoreKeepsSnapshot(t *testing.T) {
	m := New()
	plan := d.Record{ID: "plan-1", Name: "pilot", Segments: []d.Segment{{ID: "jp", Members: []string{"u-1"}}}}
	if err := m.SaveTargetingPlan(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	plan.Segments[0].Members[0] = "u-9"
	stored, err := m.Load(context.Background(), "plan-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Segments[0].Members[0] != "u-1" {
		t.Fatalf("stored snapshot was mutated: %#v", stored)
	}
}
