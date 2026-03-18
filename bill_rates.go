package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// BillRate represents a bill rate assignment in Mosaic.
// Field names correspond to the snake_case JSON fields returned by the Mosaic API.
type BillRate struct {
	// MosaicID is the unique identifier for the bill rate in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this bill rate belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// MemberID is the ID of the member this bill rate applies to.
	MemberID int `json:"member_id,omitempty"`

	// RateID is the ID of the rate associated with this bill rate.
	RateID int `json:"rate_id,omitempty"`

	// RateAmount is the billing rate amount.
	RateAmount float64 `json:"rate_amount,omitempty"`

	// RateDescription is the description of the billing rate.
	RateDescription string `json:"rate_description,omitempty"`

	// StartDate is the effective start date for this bill rate.
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the effective end date for this bill rate.
	EndDate string `json:"end_date,omitempty"`

	// Override indicates whether this bill rate overrides the default.
	Override *bool `json:"override,omitempty"`

	// OverwriteAllBillRates indicates whether to overwrite all existing bill rates.
	OverwriteAllBillRates *bool `json:"overwrite_all_bill_rates,omitempty"`

	// CreatedAt is the timestamp when the bill rate was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the bill rate was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// BillRateFilter contains filter parameters for listing bill rates.
type BillRateFilter struct {
	// IncludeArchived includes archived bill rates when set to true.
	IncludeArchived *bool
}

// toQuery converts BillRateFilter into URL query parameter values.
func (f BillRateFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.IncludeArchived != nil {
		q.Set("include_archived", strconv.FormatBool(*f.IncludeArchived))
	}
	return q
}

// billRateListResponse is the wrapper for the GET /api/{team_id}/bill_rate response,
// which returns {"team_rates": [...]}.
type billRateListResponse struct {
	TeamRates []BillRate `json:"team_rates"`
}

// ListBillRates retrieves bill rates matching the given filter.
// Uses GET /api/{team_id}/bill_rate.
func (c *Client) ListBillRates(ctx context.Context, filter BillRateFilter) ([]BillRate, error) {
	var resp billRateListResponse
	path := c.apiPath("bill_rate")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing bill rates: %w", err)
	}
	return resp.TeamRates, nil
}

// CreateBillRate creates a new bill rate and returns the created bill rate.
// Uses POST /api/{team_id}/bill_rate.
func (c *Client) CreateBillRate(ctx context.Context, billRate *BillRate) (*BillRate, error) {
	var created BillRate
	path := c.apiPath("bill_rate")
	if err := c.post(ctx, path, billRate, &created); err != nil {
		return nil, fmt.Errorf("creating bill rate: %w", err)
	}
	return &created, nil
}

// UpdateBillRate updates an existing bill rate and returns the updated bill rate.
// Uses PUT /api/{team_id}/bill_rate/{bill_rate_id}.
func (c *Client) UpdateBillRate(ctx context.Context, id int, billRate *BillRate) (*BillRate, error) {
	var updated BillRate
	path := c.apiPath("bill_rate", strconv.Itoa(id))
	if err := c.put(ctx, path, billRate, &updated); err != nil {
		return nil, fmt.Errorf("updating bill rate %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteBillRate deletes a bill rate by ID.
// Uses DELETE /api/{team_id}/bill_rate/{bill_rate_id}.
func (c *Client) DeleteBillRate(ctx context.Context, id int) error {
	path := c.apiPath("bill_rate", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting bill rate %d: %w", id, err)
	}
	return nil
}
