package mosaic

import (
	"context"
	"fmt"
	"strconv"
)

// CostRate represents a cost rate assignment in Mosaic.
// Field names correspond to the snake_case JSON fields returned by the Mosaic API.
type CostRate struct {
	// MosaicID is the unique identifier for the cost rate in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this cost rate belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// MemberID is the ID of the member this cost rate applies to.
	MemberID int `json:"member_id,omitempty"`

	// ActiveRateType is the active rate type for this cost rate.
	ActiveRateType string `json:"active_rate_type,omitempty"`

	// ActualHourlyRate is the actual hourly cost rate.
	ActualHourlyRate float64 `json:"actual_hourly_rate,omitempty"`

	// AnnualRate is the annual cost rate.
	AnnualRate float64 `json:"annual_rate,omitempty"`

	// HourlyRate is the hourly cost rate.
	HourlyRate float64 `json:"hourly_rate,omitempty"`

	// HoursPerYear is the number of working hours per year.
	HoursPerYear float64 `json:"hours_per_year,omitempty"`

	// OverheadFactor is the overhead factor applied to the cost rate.
	OverheadFactor float64 `json:"overhead_factor,omitempty"`

	// CurrencyCode is the currency code for this cost rate (e.g. "USD").
	CurrencyCode string `json:"currency_code,omitempty"`

	// Description is the cost rate description.
	Description string `json:"description,omitempty"`

	// StartDate is the effective start date for this cost rate.
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the effective end date for this cost rate.
	EndDate string `json:"end_date,omitempty"`

	// CreatedAt is the timestamp when the cost rate was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the cost rate was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// costRateListResponse is the wrapper for the GET /api/{team_id}/cost_rate response,
// which returns {"cost_rates": [...]}.
type costRateListResponse struct {
	CostRates []CostRate `json:"cost_rates"`
}

// ListCostRates retrieves all cost rates.
// Uses GET /api/{team_id}/cost_rate.
func (c *Client) ListCostRates(ctx context.Context) ([]CostRate, error) {
	var resp costRateListResponse
	path := c.apiPath("cost_rate")
	if err := c.get(ctx, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing cost rates: %w", err)
	}
	return resp.CostRates, nil
}

// CreateCostRate creates a new cost rate and returns the created cost rate.
// Uses POST /api/{team_id}/cost_rate.
func (c *Client) CreateCostRate(ctx context.Context, costRate *CostRate) (*CostRate, error) {
	var created CostRate
	path := c.apiPath("cost_rate")
	if err := c.post(ctx, path, costRate, &created); err != nil {
		return nil, fmt.Errorf("creating cost rate: %w", err)
	}
	return &created, nil
}

// UpdateCostRate updates an existing cost rate and returns the updated cost rate.
// Uses PUT /api/{team_id}/cost_rate/{cost_rate_id}.
func (c *Client) UpdateCostRate(ctx context.Context, id int, costRate *CostRate) (*CostRate, error) {
	var updated CostRate
	path := c.apiPath("cost_rate", strconv.Itoa(id))
	if err := c.put(ctx, path, costRate, &updated); err != nil {
		return nil, fmt.Errorf("updating cost rate %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteCostRate deletes a cost rate by ID.
// Uses DELETE /api/{team_id}/cost_rate/{cost_rate_id}.
func (c *Client) DeleteCostRate(ctx context.Context, id int) error {
	path := c.apiPath("cost_rate", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting cost rate %d: %w", id, err)
	}
	return nil
}
