package infrastructure

import (
	"context"
	"errors"
	"testing"
)

func TestGovernanceAuditWrapsFailure(t *testing.T) {
	auditErr := errors.New("audit sink unavailable")
	err := New().AppendGovernanceAudit(context.Background(), "release denied", func(context.Context, string) error { return auditErr })
	if !errors.Is(err, auditErr) {
		t.Fatalf("audit error identity was lost: %v", err)
	}
}
