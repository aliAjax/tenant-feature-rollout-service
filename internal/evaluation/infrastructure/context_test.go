package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEvaluationStoreHonorsContext(t *testing.T) {
	gate := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err := NewWithGate(gate).LoadEvaluationContext(ctx, "blocked")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error, got %v", err)
	}
	if time.Since(started) > 200*time.Millisecond {
		t.Fatalf("store ignored deadline for %s", time.Since(started))
	}
}
