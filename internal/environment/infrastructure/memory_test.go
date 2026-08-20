package infrastructure

import (
	"context"
	"errors"
	"testing"
)

func TestEnvironmentStoreTypedNil(t *testing.T) {
	e, err := New().LoadEnvironment(context.Background(), "missing")
	if e != nil || !errors.Is(err, ErrEnvironmentNotFound) {
		t.Fatalf("expected nil and not-found, got env=%#v err=%v", e, err)
	}
}
