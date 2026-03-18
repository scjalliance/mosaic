package mosaic

import (
	"context"
	"fmt"
)

// TeamCurrency represents a currency configured for a team in Mosaic.
// Field names correspond to the snake_case JSON fields used by the Mosaic API.
type TeamCurrency struct {
	// MosaicID is the unique identifier for the team currency.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team this currency belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// CurrencyCode is the ISO currency code (e.g. "USD", "EUR").
	CurrencyCode string `json:"currency_code,omitempty"`

	// IsDefault indicates whether this is the team's default currency.
	IsDefault bool `json:"is_default,omitempty"`

	// ArchivedAt is the timestamp when the currency was archived, or nil if active.
	ArchivedAt *string `json:"archived_at,omitempty"`

	// CreatedAt is the timestamp when the team currency was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the team currency was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// teamCurrencyBatchRequest is the request body for batch creating or deleting team currencies.
type teamCurrencyBatchRequest struct {
	CurrencyCodes []string `json:"currency_codes"`
}

// teamCurrencyListResponse is the wrapper for the GET /api/{team_id}/team_currency response.
type teamCurrencyListResponse struct {
	TeamCurrencies []TeamCurrency `json:"team_currencies"`
}

// ListTeamCurrencies retrieves all currencies configured for the team.
// Uses GET /api/{team_id}/team_currency.
func (c *Client) ListTeamCurrencies(ctx context.Context) ([]TeamCurrency, error) {
	var resp teamCurrencyListResponse
	path := c.apiPath("team_currency")
	if err := c.get(ctx, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing team currencies: %w", err)
	}
	return resp.TeamCurrencies, nil
}

// CreateTeamCurrencies adds one or more currencies to the team.
// Uses POST /api/{team_id}/team_currency with a batch request body.
func (c *Client) CreateTeamCurrencies(ctx context.Context, currencyCodes []string) error {
	path := c.apiPath("team_currency")
	body := teamCurrencyBatchRequest{CurrencyCodes: currencyCodes}
	if err := c.post(ctx, path, body, nil); err != nil {
		return fmt.Errorf("creating team currencies: %w", err)
	}
	return nil
}

// DeleteTeamCurrencies removes one or more currencies from the team.
// Uses DELETE /api/{team_id}/team_currency with a batch request body.
func (c *Client) DeleteTeamCurrencies(ctx context.Context, currencyCodes []string) error {
	path := c.apiPath("team_currency")
	body := teamCurrencyBatchRequest{CurrencyCodes: currencyCodes}
	if err := c.deleteWithBody(ctx, path, body); err != nil {
		return fmt.Errorf("deleting team currencies: %w", err)
	}
	return nil
}
