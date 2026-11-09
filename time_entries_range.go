package mosaic

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// timeEntryFloorYear is the earliest year ListProjectTimeEntries searches
// for a project's time; early enough for any mistyped date.
const timeEntryFloorYear = 1900

// timeEntryCeiling ends the trailing range that catches entries dated after
// the current year (pre-entered or mistyped).
var timeEntryCeiling = time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)

// ListProjectTimeEntries returns every time entry of one project, fetched a
// year at a time.
//
// Mosaic's API gives up on a request after 60 seconds, and one all=true
// request for a large project (about 338,000 entries for 1232011) takes
// longer, so it fails with a 504. Mosaic asked us to send date ranges
// instead (2026-11-06). This finds the project's first year with a few
// one-entry requests, fetches each year through now's, then one range for
// anything dated later. A year that still times out is fetched month by
// month instead; a month that times out is an error. A timed-out request is never repeated, since Mosaic keeps
// running it after it gives up.
func (c *Client) ListProjectTimeEntries(ctx context.Context, projectID int, now time.Time) ([]TimeEntry, error) {
	first, ok, err := c.firstTimeEntryYear(ctx, projectID, now.Year())
	if err != nil {
		return nil, err
	}
	if !ok {
		first = now.Year() + 1 // nothing through this year; still check later dates
	}
	var all []TimeEntry
	for y := first; y <= now.Year(); y++ {
		start := time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
		entries, err := c.timeEntriesBetween(ctx, projectID, start, start.AddDate(1, 0, -1))
		if err == nil {
			all = append(all, entries...)
			continue
		}
		if !isTimeout(ctx, err) {
			return nil, fmt.Errorf("time entries for project %d in %d: %w", projectID, y, err)
		}
		for m := 0; m < 12; m++ {
			ms := start.AddDate(0, m, 0)
			entries, err := c.timeEntriesBetween(ctx, projectID, ms, ms.AddDate(0, 1, -1))
			if err != nil {
				return nil, fmt.Errorf("time entries for project %d in %s (after %d timed out): %w", projectID, ms.Format("2006-01"), y, err)
			}
			all = append(all, entries...)
		}
	}
	// Entries dated after this year: usually none, so this is quick.
	next := time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)
	later, err := c.timeEntriesBetween(ctx, projectID, next, timeEntryCeiling)
	if err != nil {
		return nil, fmt.Errorf("time entries for project %d dated after %d: %w", projectID, now.Year(), err)
	}
	return append(all, later...), nil
}

// firstTimeEntryYear finds the earliest year in [timeEntryFloorYear, last]
// with a time entry for the project, by binary search over one-entry
// requests. ok is false when the project has no time through last.
func (c *Client) firstTimeEntryYear(ctx context.Context, projectID, last int) (year int, ok bool, err error) {
	floor := time.Date(timeEntryFloorYear, 1, 1, 0, 0, 0, 0, time.UTC)
	hasThrough := func(through int) (bool, error) {
		end := time.Date(through, 12, 31, 0, 0, 0, 0, time.UTC)
		entries, err := c.ListTimeEntries(ctx, TimeEntryFilter{
			ListParams: ListParams{Limit: 1},
			ProjectIDs: []int{projectID},
			StartDate:  &Date{floor},
			EndDate:    &Date{end},
		})
		if err != nil {
			return false, fmt.Errorf("finding the first year of time for project %d: %w", projectID, err)
		}
		return len(entries) > 0, nil
	}
	has, err := hasThrough(last)
	if err != nil || !has {
		return 0, false, err
	}
	lo, hi := timeEntryFloorYear, last
	for lo < hi {
		mid := lo + (hi-lo)/2
		has, err := hasThrough(mid)
		if err != nil {
			return 0, false, err
		}
		if has {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo, true, nil
}

// timeEntriesBetween fetches all of a project's time entries dated from
// start through end, inclusive.
func (c *Client) timeEntriesBetween(ctx context.Context, projectID int, start, end time.Time) ([]TimeEntry, error) {
	all := true
	return c.ListTimeEntries(ctx, TimeEntryFilter{
		ListParams: ListParams{All: &all},
		ProjectIDs: []int{projectID},
		StartDate:  &Date{start},
		EndDate:    &Date{end},
	})
}

// isTimeout reports whether err is Mosaic giving up (HTTP 504) or our own
// HTTP client timing out, while the caller's context is still live. A
// cancelled or expired caller context is not a timeout to work around.
func isTimeout(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusGatewayTimeout {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
