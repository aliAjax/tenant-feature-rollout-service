package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	d "example.com/feature-rollout-control/internal/quota/domain"
	"example.com/feature-rollout-control/internal/quota/infrastructure"
)

type staleReservation struct{}

func (staleReservation) Reserve(context.Context, string) error {
	return infrastructure.ErrStateConflict
}
func (staleReservation) Load(context.Context, string) (d.Record, error) {
	return d.Record{ID: "q-1", Name: "production", State: d.Released, Version: 2}, nil
}

func TestQuotaHTTPReportsReservationState(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/quota/reserve?id=q-1", nil)
	Handler{Service: staleReservation{}}.ServeQuotaReservation(w, r)
	var current d.Record
	if err := json.NewDecoder(w.Body).Decode(&current); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusConflict || current.State != d.Released {
		t.Fatalf("expected released conflict response, status=%d state=%s", w.Code, current.State)
	}
}
