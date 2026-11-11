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
	const interval = 40 * time.Millisecond
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
	// Allow a little scheduling slack below the interval.
	for i := 1; i < len(starts); i++ {
		if gap := starts[i].Sub(starts[i-1]); gap < interval-5*time.Millisecond {
			t.Errorf("request %d started %s after the one before, want at least %s", i, gap, interval)
		}
	}
}

// TestRateLimitPausesAllRequests checks that a 429 holds back the next
// request for the Retry-After time, and that the retry then succeeds.
func TestRateLimitPausesAllRequests(t *testing.T) {
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
