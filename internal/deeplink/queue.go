// Package deeplink buffers protocol URLs until the frontend is ready.
package deeplink

import (
	"net/url"
	"slices"
	"strings"
	"sync"
)

// Queue is safe for concurrent launch callbacks. Its zero value is ready to use.
type Queue struct {
	mu      sync.Mutex
	pending []string
	ready   bool
}

// Receive returns links for immediate delivery, or queues them until Ready.
func (q *Queue) Receive(scheme string, args []string) []string {
	var links []string
	for _, arg := range args {
		u, err := url.Parse(arg)
		if err == nil && strings.EqualFold(u.Scheme, scheme) && !slices.Contains(links, arg) {
			links = append(links, arg)
		}
	}
	if len(links) == 0 {
		return nil
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.ready {
		return links
	}
	for _, link := range links {
		if !slices.Contains(q.pending, link) {
			q.pending = append(q.pending, link)
		}
	}
	return nil
}

// Ready marks the frontend ready and drains the pending URLs.
func (q *Queue) Ready() []string {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.ready = true
	links := q.pending
	q.pending = nil
	return links
}

// IsReady reports whether the readiness handshake has completed.
func (q *Queue) IsReady() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.ready
}
