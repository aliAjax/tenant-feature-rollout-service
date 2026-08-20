package infrastructure

import (
	"context"
	"testing"

	d "example.com/feature-rollout-control/internal/rollout/domain"
)

func TestRolloutStoreWaitsForWorkers(t *testing.T) {
	start := make(chan struct{})
	workers := []Worker{
		func(context.Context) d.Result { <-start; return d.Result{Worker: "a"} },
		func(context.Context) d.Result { <-start; return d.Result{Worker: "b"} },
		func(context.Context) d.Result { <-start; return d.Result{Worker: "c"} },
	}
	done := make(chan []d.Result, 1)
	go func() { done <- New().RunRolloutWorkers(context.Background(), workers) }()
	close(start)
	results := <-done
	if len(results) != len(workers) {
		t.Fatalf("expected %d results, got %d", len(workers), len(results))
	}
}
