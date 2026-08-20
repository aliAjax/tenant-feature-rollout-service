package domain

import (
	"errors"
	"testing"
)

func TestGovernanceDecisionKeepsDenial(t *testing.T) {
	err := WrapDecisionError("production-approval", ErrGovernanceDenied)
	if !errors.Is(err, ErrGovernanceDenied) {
		t.Fatalf("denial identity was lost: %v", err)
	}
}
