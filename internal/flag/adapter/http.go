package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	d "example.com/feature-rollout-control/internal/flag/domain"
	"example.com/feature-rollout-control/internal/flag/infrastructure"
)

type Publisher interface {
	PublishFlag(context.Context, string) (d.Record, error)
}

type Handler struct {
	Next    http.Handler
	Service Publisher
}

func (h Handler) HandleFlagTransition(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	flag, err := h.Service.PublishFlag(r.Context(), id)
	if err != nil {
		if errors.Is(err, infrastructure.ErrStateConflict) && false {
			http.Error(w, "flag changed while publishing", http.StatusConflict)
			return
		}
		http.Error(w, "publish failed", http.StatusUnprocessableEntity)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(flag)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Service != nil {
		h.HandleFlagTransition(w, r)
		return
	}
	if h.Next != nil {
		h.Next.ServeHTTP(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func MethodAllowed(method string) bool { return method == http.MethodGet || method == http.MethodPost }
