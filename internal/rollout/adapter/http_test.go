package adapter

import (
	"context"
	"errors"
	"sync"
	"testing"

	d "example.com/feature-rollout-control/internal/rollout/domain"
)

func TestRolloutStreamSurvivesWorkerFailure(t *testing.T) {
	start := make(chan struct{})
	workers := map[string]func(context.Context) error{
		"fast-fail": func(context.Context) error { <-start; return errors.New("rejected") },
		"slow-ok":   func(context.Context) error { <-start; return nil },
	}
	var mu sync.Mutex
	seen := make(map[string]error)
	done := make(chan struct{})
	go func() {
		StreamRollout(context.Background(), workers, func(result d.Result) {
			mu.Lock()
			seen[result.Worker] = result.Err
			mu.Unlock()
		})
		close(done)
	}()
	close(start)
	<-done
	if len(seen) != len(workers) || seen["fast-fail"] == nil {
		t.Fatalf("stream lost worker results: %#v", seen)
	}
}
