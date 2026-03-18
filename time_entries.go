package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// TimeEntry represents a time entry in Mosaic.
// Field names correspond to the snake_case JSON fields returned by the Mosaic API.
type TimeEntry struct {
	// MosaicID is the unique identifier for the time entry.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this time entry belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// MemberID is the ID of the member who logged the time.
	MemberID int `json:"member_id,omitempty"`

	// MemberName is the display name of the member.
	MemberName string `json:"member_name,omitempty"`

	// MemberEmail is the email address of the member.
	MemberEmail string `json:"member_email,omitempty"`

	// ProjectID is the ID of the associated project.
	ProjectID int `json:"project_id,omitempty"`

	// ProjectTitle is the title of the associated project.
	ProjectTitle string `json:"project_title,omitempty"`

	// PhaseID is the ID of the associated phase.
	PhaseID int `json:"phase_id,omitempty"`

	// PhaseName is the name of the associated phase.
	PhaseName string `json:"phase_name,omitempty"`

	// PortfolioID is the ID of the associated portfolio.
	PortfolioID int `json:"portfolio_id,omitempty"`

	// PortfolioName is the name of the associated portfolio.
	PortfolioName string `json:"portfolio_name,omitempty"`

	// ActivityPhaseID is the ID of the activity phase.
	ActivityPhaseID int `json:"activity_phase_id,omitempty"`

	// Date is the date the time was logged (YYYY-MM-DD format from the API).
	Date Date `json:"date"`

	// Hours is the number of hours as a string (e.g. "2.0").
	Hours string `json:"hours,omitempty"`

	// RecordedHours is the number of hours recorded, as a string (e.g. "2.0").
	RecordedHours string `json:"recorded_hours,omitempty"`

	// InvoiceHours is the number of hours to be invoiced, as a string. May be null.
	InvoiceHours *string `json:"invoice_hours,omitempty"`

	// Rate is the billing rate as a string (e.g. "298.04"). May be null.
	Rate *string `json:"rate,omitempty"`

	// CostRate is the cost rate as a string. May be null.
	CostRate *string `json:"cost_rate,omitempty"`

	// Description is the note or description for the time entry.
	Description string `json:"description,omitempty"`

	// DescriptionID is the ID of a predefined description.
	DescriptionID int `json:"description_id,omitempty"`

	// StandardWorkCategoryID is the ID of the work category (activity type).
	StandardWorkCategoryID int `json:"standard_work_category_id,omitempty"`

	// StandardWorkCategoryTitle is the title of the work category.
	StandardWorkCategoryTitle string `json:"standard_work_category_title,omitempty"`

	// Status is the approval status of the time entry (e.g. "not_submitted").
	Status string `json:"status,omitempty"`

	// IsBillable indicates whether this time entry is billable.
	IsBillable bool `json:"is_billable,omitempty"`

	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// timeEntryListResponse is the wrapper struct for the list time entries API response.
// The API returns {"time_entry": [...]}.
type timeEntryListResponse struct {
	TimeEntry []TimeEntry `json:"time_entry"`
}

// TimeEntryFilter contains filter parameters for listing time entries.
type TimeEntryFilter struct {
	ListParams

	// MemberIDs filters time entries by one or more member IDs.
	MemberIDs []int

	// ProjectIDs filters time entries by one or more project IDs.
	ProjectIDs []int

	// StartDate filters time entries on or after this date (MM/DD/YYYY).
	StartDate *Date

	// EndDate filters time entries on or before this date (MM/DD/YYYY).
	EndDate *Date

	// TimeEntryIDs filters to specific time entry IDs.
	TimeEntryIDs []int

	// PhaseNames filters by phase names.
	PhaseNames []string

	// PhaseNumbers filters by phase numbers.
	PhaseNumbers []string

	// Clients filters by client names.
	Clients []string

	// StandardWorkCategoryIDs filters by work category IDs.
	StandardWorkCategoryIDs []int

	// StatusIDs filters by status IDs.
	StatusIDs []int

	// SyncStatuses filters by sync statuses.
	SyncStatuses []string

	// Billable filters by billable status.
	Billable *bool

	// CreatedByIntegrations filters to entries created by integrations.
	CreatedByIntegrations *bool

	// Depth controls the depth of nested data returned.
	Depth string

	// Export triggers an export format.
	Export string
}

// toQuery converts TimeEntryFilter into URL query parameter values.
func (f TimeEntryFilter) toQuery() url.Values {
	q := f.ListParams.toQuery()
	for _, id := range f.MemberIDs {
		q.Add("member_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.ProjectIDs {
		q.Add("project_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.TimeEntryIDs {
		q.Add("time_entry_ids[]", strconv.Itoa(id))
	}
	for _, name := range f.PhaseNames {
		q.Add("phase_names[]", name)
	}
	for _, num := range f.PhaseNumbers {
		q.Add("phase_numbers[]", num)
	}
	for _, client := range f.Clients {
		q.Add("clients[]", client)
	}
	for _, id := range f.StandardWorkCategoryIDs {
		q.Add("standard_work_category_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.StatusIDs {
		q.Add("status_ids[]", strconv.Itoa(id))
	}
	for _, status := range f.SyncStatuses {
		q.Add("sync_statuses[]", status)
	}
	if f.StartDate != nil && !f.StartDate.IsZero() {
		q.Set("start_date", f.StartDate.String())
	}
	if f.EndDate != nil && !f.EndDate.IsZero() {
		q.Set("end_date", f.EndDate.String())
	}
	if f.Billable != nil {
		q.Set("billable", strconv.FormatBool(*f.Billable))
	}
	if f.CreatedByIntegrations != nil {
		q.Set("created_by_integrations", strconv.FormatBool(*f.CreatedByIntegrations))
	}
	if f.Depth != "" {
		q.Set("depth", f.Depth)
	}
	if f.Export != "" {
		q.Set("export", f.Export)
	}
	return q
}

// ListTimeEntries retrieves a list of time entries matching the given filter.
// Uses GET /api/{team_id}/time_entry. The response is wrapped in {"time_entry": [...]}.
func (c *Client) ListTimeEntries(ctx context.Context, filter TimeEntryFilter) ([]TimeEntry, error) {
	var resp timeEntryListResponse
	path := c.apiPath("time_entry")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing time entries: %w", err)
	}
	return resp.TimeEntry, nil
}

// CreateTimeEntry creates a new time entry and returns the created entry.
// Uses POST /api/{team_id}/time_entry.
func (c *Client) CreateTimeEntry(ctx context.Context, entry *TimeEntry) (*TimeEntry, error) {
	var created TimeEntry
	path := c.apiPath("time_entry")
	if err := c.post(ctx, path, entry, &created); err != nil {
		return nil, fmt.Errorf("creating time entry: %w", err)
	}
	return &created, nil
}

// UpdateTimeEntry updates an existing time entry and returns the updated entry.
// Uses PUT /api/{team_id}/time_entry (body, not path param).
func (c *Client) UpdateTimeEntry(ctx context.Context, entry *TimeEntry) (*TimeEntry, error) {
	var updated TimeEntry
	path := c.apiPath("time_entry")
	if err := c.put(ctx, path, entry, &updated); err != nil {
		return nil, fmt.Errorf("updating time entry: %w", err)
	}
	return &updated, nil
}

// timeEntryDeleteRequest is the request body for deleting time entries.
type timeEntryDeleteRequest struct {
	TimeEntryIDs []int `json:"time_entry_ids"`
}

// DeleteTimeEntries deletes one or more time entries by their IDs.
// Uses DELETE /api/{team_id}/time_entry with a body containing time_entry_ids array.
func (c *Client) DeleteTimeEntries(ctx context.Context, ids []int) error {
	path := c.apiPath("time_entry")
	body := timeEntryDeleteRequest{TimeEntryIDs: ids}
	if err := c.deleteWithBody(ctx, path, body); err != nil {
		return fmt.Errorf("deleting time entries: %w", err)
	}
	return nil
}
