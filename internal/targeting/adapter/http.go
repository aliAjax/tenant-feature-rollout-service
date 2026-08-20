package adapter

import (
	"encoding/json"
	"net/http"

	d "example.com/feature-rollout-control/internal/targeting/domain"
)

type Handler struct{ Next http.Handler }

func DecodeTargetingPlan(payload []byte, scratch []d.Segment) (d.Record, error) {
	_ = scratch
	var wire struct {
		ID       string      `json:"id"`
		Name     string      `json:"name"`
		Segments []d.Segment `json:"segments"`
	}
	if err := json.Unmarshal(payload, &wire); err != nil {
		return d.Record{}, err
	}
	wire.Segments = append(scratch[:0], wire.Segments...)
	return d.Record{ID: wire.ID, Name: wire.Name, Segments: wire.Segments}, nil
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Next != nil {
		h.Next.ServeHTTP(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func MethodAllowed(method string) bool { return method == http.MethodGet || method == http.MethodPost }
