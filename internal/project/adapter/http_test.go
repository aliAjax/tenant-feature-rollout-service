package adapter

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	d "example.com/feature-rollout-control/internal/project/domain"
)

type conflictRegistrar struct{}

func (conflictRegistrar) RegisterProject(context.Context, d.Project) error {
	return d.ErrProjectConflict
}

func TestProjectHTTPReturnsConflict(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`{"ID":"p-2","TenantID":"acme","Name":"checkout"}`))
	w := httptest.NewRecorder()
	Handler{Service: conflictRegistrar{}}.ServeProjectHTTP(w, r)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}
