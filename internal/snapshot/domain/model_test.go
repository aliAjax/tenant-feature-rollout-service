package domain

import (
	"errors"
	"testing"
)

func TestSnapshotLeasePreservesPrimaryError(t *testing.T) {
	primary := errors.New("write snapshot")
	closeErr := errors.New("close lease")
	err := CloseSnapshotLease(primary, func() error { return closeErr })
	if !errors.Is(err, primary) || !errors.Is(err, closeErr) {
		t.Fatalf("expected both errors, got %v", err)
	}
}
