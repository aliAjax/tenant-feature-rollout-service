package adapter

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type capturingBatch struct{ seen error }

func (c *capturingBatch) EvaluateBatchWithContext(ctx context.Context, _ []string, _ func(context.Context, string) error) error {
	c.seen = ctx.Err()
	return c.seen
}

func TestEvaluationHTTPPropagatesCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := httptest.NewRequest(http.MethodPost, "/evaluate", bytes.NewBufferString(`{"keys":["flag-a"]}`)).WithContext(ctx)
	w := httptest.NewRecorder()
	service := &capturingBatch{}
	Handler{Service: service, Run: func(context.Context, string) error { return nil }}.ServeEvaluationHTTP(w, r)
	if service.seen != context.Canceled || w.Code != http.StatusRequestTimeout {
		t.Fatalf("context was not propagated: seen=%v status=%d", service.seen, w.Code)
	}
}
