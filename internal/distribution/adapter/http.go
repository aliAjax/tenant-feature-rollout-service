package adapter

import (
	"context"
	"encoding/json"
	"net/http"

	d "example.com/feature-rollout-control/internal/distribution/domain"
)

type Handler struct{ Next http.Handler }

func ServeDistributionSnapshot(ctx context.Context, load func(context.Context) (d.Bundle, error), encode func(d.Bundle) error) error {
	bundle, err := load(ctx)
	if err != nil {
		return err
	}
	return encode(bundle)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Next != nil {
		h.Next.ServeHTTP(w, r)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func MethodAllowed(method string) bool { return method == http.MethodGet || method == http.MethodPost }
