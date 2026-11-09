package mosaic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// fakeTimeEntries serves /api/team/time_entry from a fixed list of entry
// dates, honoring start_date, end_date and limit, and answering 504 for any
// request whose range is wider than maxDays.
type fakeTimeEntries struct {
	mu      sync.Mutex
	dates   []string // YYYY-MM-DD
	maxDays int
	ranges  []string
}

func (f *fakeTimeEntries) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	start, _ := time.Parse(dateFormat, q.Get("start_date"))
	end, _ := time.Parse(dateFormat, q.Get("end_date"))
	f.mu.Lock()
	f.ranges = append(f.ranges, q.Get("start_date")+".."+q.Get("end_date")+" limit="+q.Get("limit"))
	f.mu.Unlock()
	if q.Get("limit") != "1" && f.maxDays > 0 && end.Sub(start) > time.Duration(f.maxDays)*24*time.Hour {
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte(`{"message":"Mosaic did not answer within 60 seconds"}`))
		return
	}
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
	json.NewEncoder(w).Encode(map[string]any{"time_entry": out})
}

func TestListProjectTimeEntriesByYear(t *testing.T) {
	f := &fakeTimeEntries{dates: []string{"2019-03-04", "2019-11-30", "2021-06-01", "2024-02-29", "2026-11-02"}}
	srv := httptest.NewServer(f)
	defer srv.Close()
	c := NewClient("k", "team", WithBaseURL(srv.URL))
	got, err := c.ListProjectTimeEntries(context.Background(), 1, time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 {
		t.Fatalf("got %d entries, want 5", len(got))
	}
	for _, r := range f.ranges {
		if r == "01/01/2000..12/31/2026 limit=" { // a whole-history fetch is what timed out
			t.Fatalf("made a whole-history request: %v", f.ranges)
		}
	}
}

func TestListProjectTimeEntriesMonthFallback(t *testing.T) {
	f := &fakeTimeEntries{dates: []string{"2024-01-15", "2024-07-04", "2024-12-31", "2025-05-05"}, maxDays: 40}
	srv := httptest.NewServer(f)
	defer srv.Close()
	c := NewClient("k", "team", WithBaseURL(srv.URL))
	got, err := c.ListProjectTimeEntries(context.Background(), 1, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d entries, want 4 (via monthly fallback)", len(got))
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
	f := &fakeTimeEntries{dates: []string{"2025-05-05"}, maxDays: 1}
	srv := httptest.NewServer(f)
	defer srv.Close()
	c := NewClient("k", "team", WithBaseURL(srv.URL))
	if _, err := c.ListProjectTimeEntries(context.Background(), 1, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("want an error when a single month still times out")
	}
}
