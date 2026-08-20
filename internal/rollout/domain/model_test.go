package domain

import (
	"sync"
	"testing"
)

func TestRolloutProgressClosesOnce(t *testing.T) {
	stream := NewProgressStream(2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			<-start
			stream.Close()
		}()
	}
	close(start)
	wg.Wait()
	if _, open := <-stream.Results(); open {
		t.Fatal("progress stream must be closed")
	}
}
