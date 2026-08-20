package application

import (
	"context"
	"sync"
	"testing"

	d "example.com/feature-rollout-control/internal/distribution/domain"
	"example.com/feature-rollout-control/internal/distribution/infrastructure"
)

func TestDistributionFanOutUsesSnapshot(t *testing.T) {
	m := infrastructure.New()
	if err := m.SaveBundle(context.Background(), d.Bundle{Tenant: "acme", Entries: map[string][]byte{"flag": []byte("v1")}}); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var mu sync.Mutex
	seen := make([]string, 0, 2)
	subscribers := []func(d.Bundle){
		func(bundle d.Bundle) {
			<-start
			bundle.Entries["flag"][0] = 'a'
			mu.Lock()
			seen = append(seen, string(bundle.Entries["flag"]))
			mu.Unlock()
		},
		func(bundle d.Bundle) {
			<-start
			bundle.Entries["flag"][0] = 'b'
			mu.Lock()
			seen = append(seen, string(bundle.Entries["flag"]))
			mu.Unlock()
		},
	}
	done := make(chan error, 1)
	go func() { done <- New(m).FanOutDistribution(context.Background(), "acme", subscribers) }()
	close(start)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 {
		t.Fatalf("expected two subscriber results, got %#v", seen)
	}
}
