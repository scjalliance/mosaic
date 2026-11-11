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

// rateLimitPause is how long a Client holds back its requests after a 429
// with no usable Retry-After. With maxRetries retries, one request keeps
// trying for 6 minutes, longer than Mosaic's 5-minute window.
const rateLimitPause = 2 * time.Minute

// pacer spaces out a Client's requests so they start at least interval
// apart, across all goroutines that share the Client, and holds them all back
// while a rate-limit pause is in force. An interval of zero only pauses.
type pacer struct {
	mu          sync.Mutex
	interval    time.Duration
	next        time.Time     // earliest start time for the next reserved slot
	pausedUntil time.Time     // no request starts before this
	wake        chan struct{} // closed by pause to wake every sleeping caller
}

// wait blocks until the caller's turn to send a request, or until ctx ends.
// Each call reserves its own slot, so concurrent callers queue in the order
// they arrive. A pause wakes every sleeping caller at once, and each takes a
// new slot counted from the end of the pause, in the order they wake.
func (p *pacer) wait(ctx context.Context) error {
	for {
		p.mu.Lock()
		at := time.Now()
		if at.Before(p.next) {
			at = p.next
		}
		if at.Before(p.pausedUntil) {
			at = p.pausedUntil
		}
		p.next = at.Add(p.interval)
		if p.wake == nil {
			p.wake = make(chan struct{})
		}
		wake := p.wake
		p.mu.Unlock()

		d := time.Until(at)
		if d <= 0 {
			return ctx.Err()
		}
		t := time.NewTimer(d)
		select {
		case <-t.C:
			// The slot is stale if a pause began after it was reserved (wake
			// closed; select may pick the timer even when both are ready) or
			// one is in force now. Then queue again from the pause's end.
			p.mu.Lock()
			stale := time.Now().Before(p.pausedUntil)
			select {
			case <-wake:
				stale = true
			default:
			}
			p.mu.Unlock()
			if !stale {
				return nil
			}
		case <-wake:
			t.Stop() // a pause began; queue again from its end
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		}
	}
}

// pause holds back every request that has not started yet, including callers
// already waiting, until at least d from now. It drops all reserved slots and
// wakes the waiting callers so they queue again from the end of the pause. It
// never shortens a pause.
func (p *pacer) pause(d time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	until := time.Now().Add(d)
	if !p.pausedUntil.Before(until) {
		return
	}
	p.pausedUntil = until
	p.next = until
	if p.wake != nil {
		close(p.wake)
		p.wake = nil
	}
}
