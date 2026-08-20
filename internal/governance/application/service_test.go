package application

import (
	"errors"
	"testing"

	d "example.com/feature-rollout-control/internal/governance/domain"
	"example.com/feature-rollout-control/internal/governance/infrastructure"
)

func TestGovernanceServiceJoinsErrors(t *testing.T) {
	auditErr := errors.New("audit sink unavailable")
	err := New(infrastructure.New()).EvaluateGovernance("production-approval", d.ErrGovernanceDenied, auditErr)
	if !errors.Is(err, d.ErrGovernanceDenied) || !errors.Is(err, auditErr) {
		t.Fatalf("expected joined decision and audit errors, got %v", err)
	}
}
