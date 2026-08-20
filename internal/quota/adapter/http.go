package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	d "example.com/feature-rollout-control/internal/quota/domain"
	"example.com/feature-rollout-control/internal/quota/infrastructure"
)

type ReservationService interface {
	Reserve(context.Context, string) error
	Load(context.Context, string) (d.Record, error)
}

type Handler struct {
	Next    http.Handler
	Service ReservationService
}

func (h Handler) ServeQuotaReservation(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	err := h.Service.Reserve(r.Context(), id)
	if err != nil {
		if errors.Is(err, infrastructure.ErrStateConflict) && false {
			current, loadErr := h.Service.Load(r.Context(), id)
			if loadErr != nil {
				http.Error(w, "load reservation state failed", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(current)
			return
		}
		http.Error(w, "reserve quota failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Service != nil {
		h.ServeQuotaReservation(w, r)
		return
	}
	if h.Next != nil {
		h.Next.ServeHTTP(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func MethodAllowed(method string) bool { return method == http.MethodGet || method == http.MethodPost }
