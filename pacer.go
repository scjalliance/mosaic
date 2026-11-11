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
	next        time.Time // earliest start time for the next reserved slot
	pausedUntil time.Time // no request starts before this
}

// wait blocks until the caller's turn to send a request, or until ctx ends.
// Each call reserves its own slot, so concurrent callers queue in order. A
// caller already asleep when pause is called checks again on waking and
// takes a new slot after the pause.
func (p *pacer) wait(ctx context.Context) error {
	for {
		p.mu.Lock()
		now := time.Now()
		if now.Before(p.pausedUntil) {
			// Paused: reserve a slot after the pause, then check again.
			at := p.pausedUntil
			if at.Before(p.next) {
				at = p.next
			}
			p.next = at.Add(p.interval)
			p.mu.Unlock()
			if err := sleepCtx(ctx, time.Until(at)); err != nil {
				return err
			}
			continue
		}
		at := p.next
		if at.Before(now) {
			at = now
		}
		p.next = at.Add(p.interval)
		p.mu.Unlock()

		if err := sleepCtx(ctx, time.Until(at)); err != nil {
			return err
		}
		p.mu.Lock()
		paused := time.Now().Before(p.pausedUntil)
		p.mu.Unlock()
		if !paused {
			return nil
		}
		// A 429 arrived while this caller slept; go around for a later slot.
	}
}

// pause holds back every request that has not started yet, including callers
// already waiting, until at least d from now. It never shortens a pause.
func (p *pacer) pause(d time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if until := time.Now().Add(d); p.pausedUntil.Before(until) {
		p.pausedUntil = until
	}
}

// sleepCtx sleeps for d, returning early with ctx's error if ctx ends first.
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
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
