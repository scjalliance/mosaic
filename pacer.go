package mosaic

import (
	"context"
	"sync"
	"time"
)

// DefaultRequestInterval is the least time between the starts of two
// requests from one Client.
//
// Mosaic allows 500 requests per 5 minutes per API key
// (https://readme.mosaicapp.com/reference/authentication#api-rate-limit),
// which is one per 0.6 seconds. On 2026-11-10 a weekly report run sent up to
// 608 requests in one 5-minute window and got 1,147 HTTP 429s; Mosaic asked us
// to space requests out. 650ms caps a Client at about 461 per 5 minutes.
const DefaultRequestInterval = 650 * time.Millisecond

// rateLimitPause is how long every request from a Client waits after a 429
// that has no usable Retry-After. Mosaic's window is 5 minutes, so a short
// pause only spends more requests on more 429s.
const rateLimitPause = 60 * time.Second

// pacer spaces out a Client's requests so they start at least interval
// apart, across all goroutines that share the Client.
type pacer struct {
	mu       sync.Mutex
	interval time.Duration
	next     time.Time // earliest start time for the next request
}

// wait blocks until the caller's turn to send a request, or until ctx ends.
// Each call reserves its own slot, so concurrent callers queue in order.
func (p *pacer) wait(ctx context.Context) error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	now := time.Now()
	at := p.next
	if at.Before(now) {
		at = now
	}
	p.next = at.Add(p.interval)
	p.mu.Unlock()

	d := time.Until(at)
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// pause holds back every request that has not started yet until at least
// d from now. It never moves the next start time earlier.
func (p *pacer) pause(d time.Duration) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if until := time.Now().Add(d); p.next.Before(until) {
		p.next = until
	}
}
