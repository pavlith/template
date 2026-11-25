package deeplink

import (
	"reflect"
	"sync"
	"testing"
)

func TestQueueLifecycle(t *testing.T) {
	var q Queue
	if q.IsReady() {
		t.Fatal("queue is ready before handshake")
	}
	if got := q.Receive("example", []string{"--flag", "https://other", "example://a", "example://a"}); len(got) != 0 {
		t.Fatalf("delivered before readiness: %v", got)
	}
	q.Receive("example", []string{"example://a", "EXAMPLE://b"})
	if got := q.Ready(); !reflect.DeepEqual(got, []string{"example://a", "EXAMPLE://b"}) {
		t.Fatalf("queued links: %v", got)
	}
	if !q.IsReady() || len(q.Ready()) != 0 {
		t.Fatal("queue did not become ready and drain")
	}
	if got := q.Receive("example", []string{"example://a", "example://a"}); !reflect.DeepEqual(got, []string{"example://a"}) {
		t.Fatalf("live links: %v", got)
	}
}

func TestConcurrentQueue(t *testing.T) {
	var q Queue
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.Receive("example", []string{"example://a"})
		}()
	}
	wg.Wait()
	if got := q.Ready(); !reflect.DeepEqual(got, []string{"example://a"}) {
		t.Fatalf("queued links: %v", got)
	}
}
