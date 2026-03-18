package mosaic

import (
	"context"
	"fmt"
	"strconv"
)

// Rate represents a billing or cost rate in Mosaic.
// Field names correspond to the snake_case JSON fields returned by the Mosaic API.
type Rate struct {
	// MosaicID is the unique identifier for the rate in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this rate belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Rate is the rate amount (returned as a string by the API, e.g. "74.79").
	Rate string `json:"rate,omitempty"`

	// Description is the rate description.
	Description string `json:"description,omitempty"`

	// IsCustom indicates whether this is a custom rate.
	IsCustom *bool `json:"is_custom,omitempty"`

	// IsDefault indicates whether this is the default rate.
	IsDefault *bool `json:"is_default,omitempty"`

	// MultiplierLow is the low end of the rate multiplier range.
	MultiplierLow float64 `json:"multiplier_low,omitempty"`

	// MultiplierHigh is the high end of the rate multiplier range.
	MultiplierHigh float64 `json:"multiplier_high,omitempty"`

	// CurrencyCode is the currency code for this rate (e.g. "USD").
	CurrencyCode string `json:"currency_code,omitempty"`

	// IsCostRate indicates whether this rate is a cost rate.
	IsCostRate *bool `json:"is_cost_rate,omitempty"`

	// MergeRateID is the ID of the rate to merge with.
	MergeRateID int `json:"merge_rate_id,omitempty"`

	// CreatedAt is the timestamp when the rate was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the rate was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// rateListResponse is the wrapper for the GET /api/{team_id}/rate response,
// which returns {"rates": [...]}.
type rateListResponse struct {
	Rates []Rate `json:"rates"`
}

// ListRates retrieves all rates.
// Uses GET /api/{team_id}/rate.
func (c *Client) ListRates(ctx context.Context) ([]Rate, error) {
	var resp rateListResponse
	path := c.apiPath("rate")
	if err := c.get(ctx, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing rates: %w", err)
	}
	return resp.Rates, nil
}

// CreateRate creates a new rate and returns the created rate.
// Uses POST /api/{team_id}/rate.
func (c *Client) CreateRate(ctx context.Context, rate *Rate) (*Rate, error) {
	var created Rate
	path := c.apiPath("rate")
	if err := c.post(ctx, path, rate, &created); err != nil {
		return nil, fmt.Errorf("creating rate: %w", err)
	}
	return &created, nil
}

// UpdateRate updates an existing rate and returns the updated rate.
// Uses PUT /api/{team_id}/rate/{rate_id}.
func (c *Client) UpdateRate(ctx context.Context, id int, rate *Rate) (*Rate, error) {
	var updated Rate
	path := c.apiPath("rate", strconv.Itoa(id))
	if err := c.put(ctx, path, rate, &updated); err != nil {
		return nil, fmt.Errorf("updating rate %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteRate deletes a rate by ID.
// Uses DELETE /api/{team_id}/rate/{rate_id}.
func (c *Client) DeleteRate(ctx context.Context, id int) error {
	path := c.apiPath("rate", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting rate %d: %w", id, err)
	}
	return nil
}
