package adapter

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	d "example.com/feature-rollout-control/internal/governance/domain"
)

func TestGovernanceHTTPReturnsForbidden(t *testing.T) {
	err := errors.Join(d.WrapDecisionError("production-approval", d.ErrGovernanceDenied), errors.New("audit warning"))
	w := httptest.NewRecorder()
	ServeGovernanceDecision(w, err)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
