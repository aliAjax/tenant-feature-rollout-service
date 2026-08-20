package adapter

import (
	"encoding/json"
	"io"
	"net/http"

	d "example.com/feature-rollout-control/internal/environment/domain"
)

func DecodeEnvironmentRequest(r io.Reader) (d.Environment, error) {
	var e d.Environment
	if err := json.NewDecoder(r).Decode(&e); err != nil {
		return d.Environment{}, err
	}
	if e.Policies == nil {
		e.Policies = make(map[string]string)
	}
	return e, nil
}

func Health(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
