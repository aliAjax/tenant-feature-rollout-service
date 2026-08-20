package adapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	d "example.com/feature-rollout-control/internal/flag/domain"
	"example.com/feature-rollout-control/internal/flag/infrastructure"
)

type conflictingPublisher struct{}

func (conflictingPublisher) PublishFlag(context.Context, string) (d.Record, error) {
	return d.Record{}, infrastructure.ErrStateConflict
}

func TestFlagHTTPReportsTransitionConflict(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/flags/publish?id=flag-1", nil)
	Handler{Service: conflictingPublisher{}}.HandleFlagTransition(w, r)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}
