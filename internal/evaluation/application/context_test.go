package application

import (
	"context"
	"errors"
	"testing"

	"example.com/feature-rollout-control/internal/evaluation/infrastructure"
)

func TestEvaluationBatchStopsOnDeadline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := New(infrastructure.New()).EvaluateBatchWithContext(ctx, []string{"a", "b", "c"}, func(context.Context, string) error {
		calls++
		if calls == 1 {
			cancel()
		}
		return nil
	})
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("expected one call then cancellation, calls=%d err=%v", calls, err)
	}
}
