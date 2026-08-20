package adapter

import (
	"testing"

	d "example.com/feature-rollout-control/internal/targeting/domain"
)

func TestTargetingHTTPBodyDoesNotAlias(t *testing.T) {
	scratch := make([]d.Segment, 1, 4)
	scratch[0] = d.Segment{ID: "kept", Members: []string{"u-old"}}
	plan, err := DecodeTargetingPlan([]byte(`{"id":"plan-1","name":"pilot","segments":[{"id":"jp","members":["u-1"],"enabled":true}]}`), scratch)
	if err != nil {
		t.Fatal(err)
	}
	plan.Segments[0].ID = "changed"
	if scratch[0].ID != "kept" {
		t.Fatalf("decoded plan reused caller buffer: %#v", scratch)
	}
}
