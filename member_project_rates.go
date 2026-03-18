package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// MemberProjectRate represents a rate assignment for a member on a specific project in Mosaic.
type MemberProjectRate struct {
	// MosaicID is the unique identifier for the member project rate in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this member project rate belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// MemberID is the ID of the member this rate applies to.
	MemberID int `json:"member_id,omitempty"`

	// ProjectID is the ID of the project this rate applies to.
	ProjectID int `json:"project_id,omitempty"`

	// PhaseID is the ID of the phase this rate applies to.
	PhaseID int `json:"phase_id,omitempty"`

	// RateID is the ID of the rate definition.
	RateID int `json:"rate_id,omitempty"`

	// RateAmount is the monetary rate amount.
	RateAmount float64 `json:"rate_amount,omitempty"`

	// IsCostRate indicates whether this is a cost rate rather than a bill rate.
	IsCostRate *bool `json:"is_cost_rate,omitempty"`

	// StartDate is the date the rate assignment begins.
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the date the rate assignment ends.
	EndDate string `json:"end_date,omitempty"`

	// CreatedAt is the timestamp when the member project rate was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the member project rate was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// memberProjectRateListResponse is the wrapper for the GET /api/{team_id}/member_project_rate response,
// which returns {"member_project_rates": [...]}.
type memberProjectRateListResponse struct {
	MemberProjectRates []MemberProjectRate `json:"member_project_rates"`
}

// MemberProjectRateFilter contains required query parameters for listing member project rates.
// The API returns 404 without these filters.
type MemberProjectRateFilter struct {
	// MemberID is the required member ID to filter by.
	MemberID int

	// ProjectID is the required project ID to filter by.
	ProjectID int
}

// toQuery converts the filter to URL query parameters.
func (f MemberProjectRateFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.MemberID > 0 {
		q.Set("member_id", strconv.Itoa(f.MemberID))
	}
	if f.ProjectID > 0 {
		q.Set("project_id", strconv.Itoa(f.ProjectID))
	}
	return q
}

// ListMemberProjectRates retrieves member project rates matching the given filter.
// The filter's MemberID and ProjectID are required; the API returns 404 without them.
// Uses GET /api/{team_id}/member_project_rate.
func (c *Client) ListMemberProjectRates(ctx context.Context, filter MemberProjectRateFilter) ([]MemberProjectRate, error) {
	var resp memberProjectRateListResponse
	path := c.apiPath("member_project_rate")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing member project rates: %w", err)
	}
	return resp.MemberProjectRates, nil
}

// CreateMemberProjectRate creates a new member project rate and returns the created member project rate.
// Uses POST /api/{team_id}/member_project_rate.
func (c *Client) CreateMemberProjectRate(ctx context.Context, rate *MemberProjectRate) (*MemberProjectRate, error) {
	var created MemberProjectRate
	path := c.apiPath("member_project_rate")
	if err := c.post(ctx, path, rate, &created); err != nil {
		return nil, fmt.Errorf("creating member project rate: %w", err)
	}
	return &created, nil
}

// UpdateMemberProjectRate updates an existing member project rate and returns the updated member project rate.
// Uses PUT /api/{team_id}/member_project_rate/{member_project_rate_id}.
func (c *Client) UpdateMemberProjectRate(ctx context.Context, id int, rate *MemberProjectRate) (*MemberProjectRate, error) {
	var updated MemberProjectRate
	path := c.apiPath("member_project_rate", strconv.Itoa(id))
	if err := c.put(ctx, path, rate, &updated); err != nil {
		return nil, fmt.Errorf("updating member project rate %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteMemberProjectRate deletes a member project rate by ID.
// Uses DELETE /api/{team_id}/member_project_rate/{member_project_rate_id}.
func (c *Client) DeleteMemberProjectRate(ctx context.Context, id int) error {
	path := c.apiPath("member_project_rate", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting member project rate %d: %w", id, err)
	}
	return nil
}
