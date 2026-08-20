package adapter

import (
	"context"
	"net/http"
	"sync"

	d "example.com/feature-rollout-control/internal/rollout/domain"
)

type Handler struct{ Next http.Handler }

func StreamRollout(ctx context.Context, workers map[string]func(context.Context) error, emit func(d.Result)) {
	results := make(chan d.Result, len(workers))
	var wg sync.WaitGroup
	wg.Add(len(workers))
	for name, run := range workers {
		name, run := name, run
		go func() {
			defer wg.Done()
			results <- d.Result{Worker: name, Err: run(ctx)}
		}()
	}
	wg.Wait()
	close(results)
	for result := range results {
		emit(result)
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Next != nil {
		h.Next.ServeHTTP(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func MethodAllowed(method string) bool { return method == http.MethodGet || method == http.MethodPost }
