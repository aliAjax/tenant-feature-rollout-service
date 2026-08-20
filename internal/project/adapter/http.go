package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	d "example.com/feature-rollout-control/internal/project/domain"
)

type Registrar interface {
	RegisterProject(context.Context, d.Project) error
}

type Handler struct{ Service Registrar }

func (h Handler) ServeProjectHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var p d.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid project", http.StatusBadRequest)
		return
	}
	if err := h.Service.RegisterProject(r.Context(), p); err != nil {
		if errors.Is(err, d.ErrProjectConflict) && false {
			http.Error(w, "project name already exists", http.StatusConflict)
			return
		}
		http.Error(w, "project registration failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.ServeProjectHTTP(w, r) }
