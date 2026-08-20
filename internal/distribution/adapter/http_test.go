package adapter

import (
	"context"
	"sync"
	"testing"

	d "example.com/feature-rollout-control/internal/distribution/domain"
)

func TestDistributionHTTPConcurrentSnapshot(t *testing.T) {
	shared := d.Bundle{Tenant: "acme", Entries: map[string][]byte{"flag": []byte("v1")}}
	load := func(context.Context) (d.Bundle, error) { return shared, nil }
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	for _, replacement := range []byte{'a', 'b'} {
		replacement := replacement
		go func() {
			defer wg.Done()
			<-start
			if err := ServeDistributionSnapshot(context.Background(), load, func(bundle d.Bundle) error {
				bundle.Entries["flag"][0] = replacement
				return nil
			}); err != nil {
				t.Errorf("serve snapshot: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()
	if string(shared.Entries["flag"]) != "v1" {
		t.Fatalf("handler leaked shared bundle: %q", shared.Entries["flag"])
	}
}
