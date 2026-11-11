package mosaic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestPacerSpacesRequests checks that requests from one Client, including
// concurrent ones, start at least the request interval apart.
func TestPacerSpacesRequests(t *testing.T) {
	const interval = 100 * time.Millisecond
	var mu sync.Mutex
	var starts []time.Time
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		starts = append(starts, time.Now())
		mu.Unlock()
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := NewClient("k", "team", WithBaseURL(srv.URL), WithRequestInterval(interval))

	var wg sync.WaitGroup
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.get(context.Background(), "/x", nil, nil); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	if len(starts) != 6 {
		t.Fatalf("got %d requests, want 6", len(starts))
	}
	// Server-side arrival times jitter, so allow 40ms of slack.
	for i := 1; i < len(starts); i++ {
		if gap := starts[i].Sub(starts[i-1]); gap < interval-40*time.Millisecond {
			t.Errorf("request %d started %s after the one before, want at least %s", i, gap, interval)
		}
	}
}

// TestRateLimitRetriesAfterPause checks that a 429 holds back the retry for
// the Retry-After time, and that the retry then succeeds.
func TestRateLimitRetriesAfterPause(t *testing.T) {
	var calls atomic.Int32
	var mu sync.Mutex
	var starts []time.Time
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		starts = append(starts, time.Now())
		mu.Unlock()
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := NewClient("k", "team", WithBaseURL(srv.URL), WithRequestInterval(time.Millisecond))

	if err := c.get(context.Background(), "/x", nil, nil); err != nil {
		t.Fatalf("get after one 429: %v", err)
	}
	if err := c.get(context.Background(), "/y", nil, nil); err != nil {
		t.Fatal(err)
	}
	if len(starts) != 3 {
		t.Fatalf("got %d requests, want 3", len(starts))
	}
	if gap := starts[1].Sub(starts[0]); gap < 2*time.Second-50*time.Millisecond {
		t.Errorf("retry started %s after the 429, want at least 2s", gap)
	}
}

// TestPacerWaitHonorsContext checks that a request waiting for its turn
// returns when the caller's context ends.
func TestPacerWaitHonorsContext(t *testing.T) {
	p := &pacer{interval: time.Hour}
	if err := p.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := p.wait(ctx); err == nil {
		t.Fatal("wait returned nil, want the context's error")
	}
}

// TestPauseHoldsBackWaitingCallers checks that a pause also holds back a
// caller that reserved its slot before the pause and is already asleep.
func TestPauseHoldsBackWaitingCallers(t *testing.T) {
	p := &pacer{interval: 100 * time.Millisecond}
	if err := p.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	done := make(chan time.Time)
	go func() {
		if err := p.wait(context.Background()); err != nil {
			t.Error(err)
		}
		done <- time.Now()
	}()
	time.Sleep(20 * time.Millisecond) // the caller is now asleep on its slot
	p.pause(300 * time.Millisecond)
	// The pause began about 20ms in and lasts 300ms. The caller should start
	// as soon as it ends, not an interval later.
	got := (<-done).Sub(start)
	if got < 300*time.Millisecond || got > 500*time.Millisecond {
		t.Errorf("waiting caller started %s after start, want about 320ms", got)
	}
}

// TestParseRetryAfter checks the Retry-After forms Mosaic may send.
func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		in   string
		want time.Duration
	}{
		{"", rateLimitPause},
		{"junk", rateLimitPause},
		{"-3", rateLimitPause},
		{"0", minRetryDelay},
		{"1", minRetryDelay},
		{"30", 30 * time.Second},
		{"Wed, 11 Nov 2026 10:00:05 GMT", 5 * time.Second},
		{"Wed, 11 Nov 2026 10:02:00 GMT", 2 * time.Minute},
		{"Wed, 11 Nov 2026 09:59:00 GMT", minRetryDelay},
		{"86400", maxRetryAfter},
		{"99999999999999999", maxRetryAfter},
		{"99999999999999999999", maxRetryAfter},
		{"-99999999999999999999", rateLimitPause},
		{"Thu, 12 Nov 2026 10:00:00 GMT", maxRetryAfter},
	} {
		if got := parseRetryAfter(tc.in, now); got != tc.want {
			t.Errorf("parseRetryAfter(%q) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

// TestPauseWakesLaterSlots checks that a caller asleep on a slot that falls
// after the pause ends is woken and starts at the end of the pause, not at
// its old slot.
func TestPauseWakesLaterSlots(t *testing.T) {
	p := &pacer{interval: 500 * time.Millisecond}
	if err := p.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	done := make(chan time.Time)
	go func() {
		if err := p.wait(context.Background()); err != nil {
			t.Error(err)
		}
		done <- time.Now()
	}()
	time.Sleep(20 * time.Millisecond) // asleep on its slot at about 500ms
	p.pause(100 * time.Millisecond)
	if got := (<-done).Sub(start); got < 100*time.Millisecond || got > 350*time.Millisecond {
		t.Errorf("caller started %s after start, want about 120ms (the end of the pause)", got)
	}
}

// TestDefaultHTTPClientTimeouts checks that the default client waits longer
// than Mosaic's 60-second limit for headers, and does not cut off a slow body.
func TestDefaultHTTPClientTimeouts(t *testing.T) {
	hc := newDefaultHTTPClient()
	tr, ok := hc.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport is %T, want *http.Transport", hc.Transport)
	}
	if tr.ResponseHeaderTimeout != 75*time.Second {
		t.Errorf("ResponseHeaderTimeout = %s, want 75s (past Mosaic's 60s)", tr.ResponseHeaderTimeout)
	}
	if hc.Timeout != 10*time.Minute {
		t.Errorf("Timeout = %s, want 10m", hc.Timeout)
	}
}

// TestDefaultHTTPClientKeepsReplacedTransport checks that a RoundTripper a
// program installed as http.DefaultTransport is used, not bypassed.
func TestDefaultHTTPClientKeepsReplacedTransport(t *testing.T) {
	orig := http.DefaultTransport
	defer func() { http.DefaultTransport = orig }()
	custom := roundTripFunc(func(r *http.Request) (*http.Response, error) { return nil, nil })
	http.DefaultTransport = custom
	hc := newDefaultHTTPClient()
	if _, ok := hc.Transport.(roundTripFunc); !ok {
		t.Errorf("transport is %T, want the installed RoundTripper", hc.Transport)
	}
	if hc.Timeout != 10*time.Minute {
		t.Errorf("Timeout = %s, want 10m", hc.Timeout)
	}
}

// roundTripFunc adapts a function to http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip calls f.
func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
