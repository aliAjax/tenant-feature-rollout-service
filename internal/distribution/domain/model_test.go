package domain

import (
	"sync"
	"testing"
)

func TestDistributionBundleDeepCopy(t *testing.T) {
	original := Bundle{Tenant: "acme", Entries: map[string][]byte{"flag": []byte("v1")}}
	clone := CloneDistributionBundle(original)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); <-start; clone.Entries["flag"][0] = 'x' }()
	go func() { defer wg.Done(); <-start; _ = original.Entries["flag"][0] }()
	close(start)
	wg.Wait()
	if string(original.Entries["flag"]) != "v1" {
		t.Fatalf("clone changed original: %q", original.Entries["flag"])
	}
}
