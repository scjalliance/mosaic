package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// TimeTracking represents a time tracking entry in Mosaic.
// Field names correspond to the snake_case JSON fields used by the Mosaic API.
type TimeTracking struct {
	// MosaicID is the unique identifier for the time tracking entry.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team this time tracking entry belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Date is the date of the time tracking entry.
	Date string `json:"date,omitempty"`

	// MemberID is the ID of the team member who logged the time.
	MemberID int `json:"member_id,omitempty"`

	// ProjectID is the ID of the associated project.
	ProjectID int `json:"project_id,omitempty"`

	// PhaseID is the ID of the associated phase.
	PhaseID int `json:"phase_id,omitempty"`

	// StandardWorkCategoryID is the associated standard work category.
	StandardWorkCategoryID int `json:"standard_work_category_id,omitempty"`

	// Description is a description of the work performed.
	Description string `json:"description,omitempty"`

	// EstimatedHours is the number of hours tracked.
	EstimatedHours float64 `json:"estimated_hours,omitempty"`

	// CreatedAt is the timestamp when the entry was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the entry was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// TimeTrackingFilter contains filter parameters for listing time tracking entries.
type TimeTrackingFilter struct {
	// Date is the date to filter by (required).
	Date string

	// MemberIDs is a comma-separated list of member IDs to filter by.
	MemberIDs string
}

// toQuery converts TimeTrackingFilter into URL query parameter values.
func (f TimeTrackingFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.Date != "" {
		q.Set("date", f.Date)
	}
	if f.MemberIDs != "" {
		q.Set("member_ids", f.MemberIDs)
	}
	return q
}

// timeTrackingListResponse is the wrapper for the GET /api/{team_id}/time_tracking response.
type timeTrackingListResponse struct {
	TimeTracking []TimeTracking `json:"time_tracking"`
}

// ListTimeTracking retrieves time tracking entries matching the given filter.
// The Date field in the filter is required by the API.
// Uses GET /api/{team_id}/time_tracking.
func (c *Client) ListTimeTracking(ctx context.Context, filter TimeTrackingFilter) ([]TimeTracking, error) {
	var resp timeTrackingListResponse
	path := c.apiPath("time_tracking")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing time tracking entries: %w", err)
	}
	return resp.TimeTracking, nil
}

// CreateTimeTracking creates a new time tracking entry and returns the created entry.
// Uses POST /api/{team_id}/time_tracking.
func (c *Client) CreateTimeTracking(ctx context.Context, entry *TimeTracking) (*TimeTracking, error) {
	var created TimeTracking
	path := c.apiPath("time_tracking")
	if err := c.post(ctx, path, entry, &created); err != nil {
		return nil, fmt.Errorf("creating time tracking entry: %w", err)
	}
	return &created, nil
}

// UpdateTimeTracking updates an existing time tracking entry and returns the updated entry.
// Uses PUT /api/{team_id}/time_tracking/{time_tracking_id}.
func (c *Client) UpdateTimeTracking(ctx context.Context, id int, entry *TimeTracking) (*TimeTracking, error) {
	var updated TimeTracking
	path := c.apiPath("time_tracking", strconv.Itoa(id))
	if err := c.put(ctx, path, entry, &updated); err != nil {
		return nil, fmt.Errorf("updating time tracking entry %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteTimeTracking deletes a time tracking entry by ID.
// Uses DELETE /api/{team_id}/time_tracking/{time_tracking_id}.
func (c *Client) DeleteTimeTracking(ctx context.Context, id int) error {
	path := c.apiPath("time_tracking", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting time tracking entry %d: %w", id, err)
	}
	return nil
}
