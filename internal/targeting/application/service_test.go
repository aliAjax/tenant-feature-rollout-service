package application

import (
	"testing"

	d "example.com/feature-rollout-control/internal/targeting/domain"
	"example.com/feature-rollout-control/internal/targeting/infrastructure"
)

func TestTargetingCompileLeavesInputUntouched(t *testing.T) {
	input := []d.Segment{
		{ID: "jp", Members: []string{"u-1"}, Enabled: true},
		{ID: "us", Members: []string{"u-2"}, Enabled: false},
	}
	out := New(infrastructure.New()).CompileTargetingPlan(input)
	out[0].Members[0] = "changed"
	if len(input) != 2 || input[0].Members[0] != "u-1" || input[1].ID != "us" {
		t.Fatalf("compile mutated input: %#v", input)
	}
}
