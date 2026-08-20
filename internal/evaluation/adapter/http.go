package adapter

import (
	"context"
	"encoding/json"
	"net/http"
)

type BatchEvaluator interface {
	EvaluateBatchWithContext(context.Context, []string, func(context.Context, string) error) error
}

type Handler struct {
	Next    http.Handler
	Service BatchEvaluator
	Run     func(context.Context, string) error
}

func (h Handler) ServeEvaluationHTTP(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Keys []string `json:"keys"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid evaluation request", http.StatusBadRequest)
		return
	}
	if err := h.Service.EvaluateBatchWithContext(context.Background(), input.Keys, h.Run); err != nil {
		http.Error(w, err.Error(), http.StatusRequestTimeout)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Service != nil {
		h.ServeEvaluationHTTP(w, r)
		return
	}
	if h.Next != nil {
		h.Next.ServeHTTP(w, r)
		return
	}
	w.WriteHeader(http.StatusNotImplemented)
}

func MethodAllowed(method string) bool {
	return method == http.MethodGet || method == http.MethodPost || method == http.MethodPut
}
