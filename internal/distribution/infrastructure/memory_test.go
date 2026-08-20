package infrastructure

import (
	"context"
	"sync"
	"testing"

	d "example.com/feature-rollout-control/internal/distribution/domain"
)

func TestDistributionStoreSnapshotIsolation(t *testing.T) {
	m := New()
	if err := m.SaveBundle(context.Background(), d.Bundle{Tenant: "acme", Version: 1, Entries: map[string][]byte{"flag": []byte("v1")}}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := m.SnapshotDistribution(context.Background(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); <-start; snapshot.Entries["flag"][0] = 'x' }()
	go func() {
		defer wg.Done()
		<-start
		_ = m.UpdateEntry(context.Background(), "acme", "flag", []byte("v2"))
	}()
	close(start)
	wg.Wait()
	latest, _ := m.SnapshotDistribution(context.Background(), "acme")
	if string(latest.Entries["flag"]) != "v2" {
		t.Fatalf("snapshot mutation reached store: %q", latest.Entries["flag"])
	}
}
