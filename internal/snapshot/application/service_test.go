package application

import (
	"context"
	"io"
	"sync"
	"testing"

	"example.com/feature-rollout-control/internal/snapshot/infrastructure"
)

type trackedFactory struct {
	mu      sync.Mutex
	open    int
	maxOpen int
}

type trackedWriter struct{ owner *trackedFactory }

func (w *trackedWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w *trackedWriter) Close() error {
	w.owner.mu.Lock()
	w.owner.open--
	w.owner.mu.Unlock()
	return nil
}

func (f *trackedFactory) Open(context.Context, string) (io.WriteCloser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.open++
	if f.open > f.maxOpen {
		f.maxOpen = f.open
	}
	return &trackedWriter{owner: f}, nil
}

func TestSnapshotBatchReleasesEachItem(t *testing.T) {
	factory := &trackedFactory{}
	err := New(infrastructure.New()).ExportSnapshots(context.Background(), []string{"one", "two", "three"}, factory, func(w io.Writer, id string) error {
		_, err := io.WriteString(w, id)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if factory.open != 0 || factory.maxOpen != 1 {
		t.Fatalf("resources accumulated: open=%d max=%d", factory.open, factory.maxOpen)
	}
}
