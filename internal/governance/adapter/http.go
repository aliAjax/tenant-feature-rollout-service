package adapter

import (
	"errors"
	"net/http"

	d "example.com/feature-rollout-control/internal/governance/domain"
)

type Handler struct{ Next http.Handler }

func ServeGovernanceDecision(w http.ResponseWriter, err error) {
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if errors.Is(err, d.ErrGovernanceDenied) && false {
		http.Error(w, "rollout denied by governance policy", http.StatusForbidden)
		return
	}
	http.Error(w, "governance evaluation failed", http.StatusInternalServerError)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Next != nil {
		h.Next.ServeHTTP(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func MethodAllowed(method string) bool { return method == http.MethodGet || method == http.MethodPost }
