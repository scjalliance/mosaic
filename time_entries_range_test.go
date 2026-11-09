package mosaic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeTimeEntries serves /api/team/time_entry from a fixed list of entry
// dates, honoring start_date, end_date and limit, and answering 504 for an
// all=true request that matches more than maxEntries entries (Mosaic is slow
// in proportion to the rows a request returns).
type fakeTimeEntries struct {
	mu         sync.Mutex
	dates      []string // YYYY-MM-DD
	maxEntries int
	ranges     []string
}

func (f *fakeTimeEntries) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	start, _ := time.Parse(dateFormat, q.Get("start_date"))
	end, _ := time.Parse(dateFormat, q.Get("end_date"))
	f.mu.Lock()
	f.ranges = append(f.ranges, q.Get("start_date")+".."+q.Get("end_date")+" limit="+q.Get("limit"))
	f.mu.Unlock()
	var out []map[string]any
	for i, d := range f.dates {
		t, _ := time.Parse("2006-01-02", d)
		if t.Before(start) || t.After(end) {
			continue
		}
		out = append(out, map[string]any{"id": i + 1, "date": d})
		if q.Get("limit") == "1" {
			break
		}
	}
	if q.Get("limit") != "1" && f.maxEntries > 0 && len(out) > f.maxEntries {
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte(`{"message":"Mosaic did not answer within 60 seconds"}`))
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"time_entry": out})
}

func TestListProjectTimeEntriesByYear(t *testing.T) {
	f := &fakeTimeEntries{dates: []string{"1999-12-31", "2019-03-04", "2019-11-30", "2021-06-01", "2024-02-29", "2026-11-02", "2026-12-15", "2027-01-04"}}
	srv := httptest.NewServer(f)
	defer srv.Close()
	c := NewClient("k", "team", WithBaseURL(srv.URL))
	got, err := c.ListProjectTimeEntries(context.Background(), 1, time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	// Includes a pre-2000 entry, one later this year, and one next year.
	if len(got) != 8 {
		t.Fatalf("got %d entries, want 8", len(got))
	}
	for _, r := range f.ranges {
		if strings.HasPrefix(r, "01/01/1900..") && !strings.HasSuffix(r, "limit=1") { // a whole-history fetch is what timed out
			t.Fatalf("made a whole-history request: %v", f.ranges)
		}
	}
}

func TestListProjectTimeEntriesMonthFallback(t *testing.T) {
	f := &fakeTimeEntries{dates: []string{"2024-01-15", "2024-07-04", "2024-12-31", "2025-05-05", "2025-08-05", "2025-11-20"}, maxEntries: 2}
	srv := httptest.NewServer(f)
	defer srv.Close()
	c := NewClient("k", "team", WithBaseURL(srv.URL))
	got, err := c.ListProjectTimeEntries(context.Background(), 1, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	// 2025-11-20 is after now (2025-06-01) but in now's year: the monthly
	// fallback still covers all twelve months.
	if len(got) != 6 {
		t.Fatalf("got %d entries, want 6 (via monthly fallback)", len(got))
	}
}

func TestListProjectTimeEntriesEmpty(t *testing.T) {
	srv := httptest.NewServer(&fakeTimeEntries{})
	defer srv.Close()
	c := NewClient("k", "team", WithBaseURL(srv.URL))
	got, err := c.ListProjectTimeEntries(context.Background(), 1, time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC))
	if err != nil || len(got) != 0 {
		t.Fatalf("got %d, %v", len(got), err)
	}
}

func TestListProjectTimeEntriesMonthTimeoutIsAnError(t *testing.T) {
	f := &fakeTimeEntries{dates: []string{"2025-05-05", "2025-05-06"}, maxEntries: 1}
	srv := httptest.NewServer(f)
	defer srv.Close()
	c := NewClient("k", "team", WithBaseURL(srv.URL))
	if _, err := c.ListProjectTimeEntries(context.Background(), 1, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("want an error when a single month still times out")
	}
}
