package application

import (
	"context"
	"errors"
	"testing"

	"example.com/feature-rollout-control/internal/rollout/infrastructure"
)

func TestRolloutCoordinatorCollectsAll(t *testing.T) {
	start := make(chan struct{})
	workers := map[string]func(context.Context) error{
		"a": func(context.Context) error { <-start; return nil },
		"b": func(context.Context) error { <-start; return errors.New("rejected") },
		"c": func(context.Context) error { <-start; return nil },
	}
	done := make(chan int, 1)
	go func() { done <- len(New(infrastructure.New()).CoordinateRollout(context.Background(), workers)) }()
	close(start)
	if count := <-done; count != len(workers) {
		t.Fatalf("expected %d results, got %d", len(workers), count)
	}
}
