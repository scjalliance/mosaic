package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// CalendarEvent represents a calendar event in Mosaic.
type CalendarEvent struct {
	// MosaicID is the unique identifier for the calendar event in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this calendar event belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Title is the calendar event's title.
	Title string `json:"title"`

	// Details is the calendar event's description or details.
	Details string `json:"details,omitempty"`

	// StartDatetime is the start date and time of the calendar event.
	StartDatetime string `json:"start_datetime,omitempty"`

	// EndDatetime is the end date and time of the calendar event.
	EndDatetime string `json:"end_datetime,omitempty"`

	// ProjectID is the ID of the project associated with the calendar event.
	ProjectID int `json:"project_id,omitempty"`

	// PhaseID is the ID of the phase associated with the calendar event.
	PhaseID int `json:"phase_id,omitempty"`

	// MemberIDs is the list of member IDs associated with the calendar event.
	MemberIDs []int `json:"member_ids,omitempty"`

	// CreatedAt is the timestamp when the calendar event was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the calendar event was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// CalendarEventFilter contains filter parameters for listing calendar events.
type CalendarEventFilter struct {
	// StartDate filters calendar events starting on or after this date.
	StartDate string

	// EndDate filters calendar events ending on or before this date.
	EndDate string

	// MemberIDs filters calendar events by member IDs.
	MemberIDs []int

	// ProjectIDs filters calendar events by project IDs.
	ProjectIDs []int

	// PhaseIDs filters calendar events by phase IDs.
	PhaseIDs []int

	// Title filters calendar events by title.
	Title string
}

// toQuery converts CalendarEventFilter into URL query parameter values.
func (f CalendarEventFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.StartDate != "" {
		q.Set("start_date", f.StartDate)
	}
	if f.EndDate != "" {
		q.Set("end_date", f.EndDate)
	}
	for _, id := range f.MemberIDs {
		q.Add("member_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.ProjectIDs {
		q.Add("project_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.PhaseIDs {
		q.Add("phase_ids[]", strconv.Itoa(id))
	}
	if f.Title != "" {
		q.Set("title", f.Title)
	}
	return q
}

// calendarEventListResponse is the wrapper for the GET /api/{team_id}/calendar_event response,
// which returns {"calendar_events": [...]}.
type calendarEventListResponse struct {
	CalendarEvents []CalendarEvent `json:"calendar_events"`
}

// ListCalendarEvents retrieves calendar events matching the given filter.
// Uses GET /api/{team_id}/calendar_event.
func (c *Client) ListCalendarEvents(ctx context.Context, filter CalendarEventFilter) ([]CalendarEvent, error) {
	var resp calendarEventListResponse
	path := c.apiPath("calendar_event")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing calendar events: %w", err)
	}
	return resp.CalendarEvents, nil
}

// CreateCalendarEvent creates a new calendar event and returns the created calendar event.
// Uses POST /api/{team_id}/calendar_event.
func (c *Client) CreateCalendarEvent(ctx context.Context, event *CalendarEvent) (*CalendarEvent, error) {
	var created CalendarEvent
	path := c.apiPath("calendar_event")
	if err := c.post(ctx, path, event, &created); err != nil {
		return nil, fmt.Errorf("creating calendar event: %w", err)
	}
	return &created, nil
}

// UpdateCalendarEvent updates an existing calendar event and returns the updated calendar event.
// Uses PUT /api/{team_id}/calendar_event/{calendar_event_id}.
func (c *Client) UpdateCalendarEvent(ctx context.Context, id int, event *CalendarEvent) (*CalendarEvent, error) {
	var updated CalendarEvent
	path := c.apiPath("calendar_event", strconv.Itoa(id))
	if err := c.put(ctx, path, event, &updated); err != nil {
		return nil, fmt.Errorf("updating calendar event %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteCalendarEvent deletes a calendar event by ID.
// Uses DELETE /api/{team_id}/calendar_event/{calendar_event_id}.
func (c *Client) DeleteCalendarEvent(ctx context.Context, id int) error {
	path := c.apiPath("calendar_event", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting calendar event %d: %w", id, err)
	}
	return nil
}
