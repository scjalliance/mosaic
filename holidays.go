package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Holiday represents a holiday in Mosaic.
type Holiday struct {
	// MosaicID is the unique identifier for the holiday in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this holiday belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Name is the display name of the holiday.
	Name string `json:"name"`

	// AllDay indicates whether the holiday spans the entire day.
	AllDay *bool `json:"all_day,omitempty"`

	// DailyHours is the number of hours for the holiday if not all day.
	DailyHours float64 `json:"daily_hours,omitempty"`

	// StartDate is the start date of the holiday (e.g. "2025-12-25").
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the end date of the holiday (e.g. "2025-12-25").
	EndDate string `json:"end_date,omitempty"`

	// CreatedAt is the timestamp when the holiday was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the holiday was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// HolidayFilter contains filter parameters for listing holidays.
type HolidayFilter struct {
	// StartDate filters holidays starting on or after this date.
	StartDate string

	// EndDate filters holidays ending on or before this date.
	EndDate string
}

// toQuery converts HolidayFilter into URL query parameter values.
func (f HolidayFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.StartDate != "" {
		q.Set("start_date", f.StartDate)
	}
	if f.EndDate != "" {
		q.Set("end_date", f.EndDate)
	}
	return q
}

// holidayListResponse is the wrapper for the GET /api/{team_id}/holidays response.
type holidayListResponse struct {
	Holidays []Holiday `json:"holidays"`
}

// ListHolidays retrieves holidays matching the given filter.
// Uses GET /api/{team_id}/holidays.
func (c *Client) ListHolidays(ctx context.Context, filter HolidayFilter) ([]Holiday, error) {
	var resp holidayListResponse
	path := c.apiPath("holidays")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing holidays: %w", err)
	}
	return resp.Holidays, nil
}

// CreateHoliday creates a new holiday and returns the created holiday.
// Uses POST /api/{team_id}/holidays.
func (c *Client) CreateHoliday(ctx context.Context, holiday *Holiday) (*Holiday, error) {
	var created Holiday
	path := c.apiPath("holidays")
	if err := c.post(ctx, path, holiday, &created); err != nil {
		return nil, fmt.Errorf("creating holiday: %w", err)
	}
	return &created, nil
}

// UpdateHoliday updates an existing holiday and returns the updated holiday.
// Uses PUT /api/{team_id}/holidays/{id}.
func (c *Client) UpdateHoliday(ctx context.Context, id int, holiday *Holiday) (*Holiday, error) {
	var updated Holiday
	path := c.apiPath("holidays", strconv.Itoa(id))
	if err := c.put(ctx, path, holiday, &updated); err != nil {
		return nil, fmt.Errorf("updating holiday %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteHoliday deletes a holiday by ID.
// Uses DELETE /api/{team_id}/holidays/{id}.
func (c *Client) DeleteHoliday(ctx context.Context, id int) error {
	path := c.apiPath("holidays", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting holiday %d: %w", id, err)
	}
	return nil
}
