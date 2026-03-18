package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// CurrencyExchangeRate represents a currency exchange rate in Mosaic.
// Field names correspond to the snake_case JSON fields used by the Mosaic API.
type CurrencyExchangeRate struct {
	// MosaicID is the unique identifier for the exchange rate.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team this exchange rate belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// SourceCurrencyCode is the ISO currency code for the source currency.
	SourceCurrencyCode string `json:"source_currency_code,omitempty"`

	// TargetCurrencyCode is the ISO currency code for the target currency.
	TargetCurrencyCode string `json:"target_currency_code,omitempty"`

	// Ratio is the exchange rate ratio from source to target currency.
	Ratio float64 `json:"ratio,omitempty"`

	// StartDate is the date the exchange rate becomes effective.
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the date the exchange rate expires.
	EndDate string `json:"end_date,omitempty"`

	// CreatedAt is the timestamp when the exchange rate was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the exchange rate was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// CurrencyExchangeRateFilter contains filter parameters for listing exchange rates.
type CurrencyExchangeRateFilter struct {
	// StartDate is required — the start of the date range.
	StartDate string

	// EndDate is required — the end of the date range.
	EndDate string
}

// toQuery converts CurrencyExchangeRateFilter into URL query parameter values.
func (f CurrencyExchangeRateFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.StartDate != "" {
		q.Set("start_date", f.StartDate)
	}
	if f.EndDate != "" {
		q.Set("end_date", f.EndDate)
	}
	return q
}

// ListCurrencyExchangeRates retrieves currency exchange rates within a date range.
// Uses GET /api/{team_id}/currency_exchange_rate.
// The API requires start_date and end_date query parameters.
func (c *Client) ListCurrencyExchangeRates(ctx context.Context, filter CurrencyExchangeRateFilter) ([]CurrencyExchangeRate, error) {
	var resp []CurrencyExchangeRate
	path := c.apiPath("currency_exchange_rate")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing currency exchange rates: %w", err)
	}
	return resp, nil
}

// CreateCurrencyExchangeRate creates a new currency exchange rate and returns the created rate.
// Uses POST /api/{team_id}/currency_exchange_rate.
func (c *Client) CreateCurrencyExchangeRate(ctx context.Context, rate *CurrencyExchangeRate) (*CurrencyExchangeRate, error) {
	var created CurrencyExchangeRate
	path := c.apiPath("currency_exchange_rate")
	if err := c.post(ctx, path, rate, &created); err != nil {
		return nil, fmt.Errorf("creating currency exchange rate: %w", err)
	}
	return &created, nil
}

// UpdateCurrencyExchangeRate updates an existing currency exchange rate and returns the updated rate.
// Uses PUT /api/{team_id}/currency_exchange_rate/{id}.
func (c *Client) UpdateCurrencyExchangeRate(ctx context.Context, id int, rate *CurrencyExchangeRate) (*CurrencyExchangeRate, error) {
	var updated CurrencyExchangeRate
	path := c.apiPath("currency_exchange_rate", strconv.Itoa(id))
	if err := c.put(ctx, path, rate, &updated); err != nil {
		return nil, fmt.Errorf("updating currency exchange rate %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteCurrencyExchangeRate deletes a currency exchange rate by ID.
// Uses DELETE /api/{team_id}/currency_exchange_rate/{id}.
func (c *Client) DeleteCurrencyExchangeRate(ctx context.Context, id int) error {
	path := c.apiPath("currency_exchange_rate", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting currency exchange rate %d: %w", id, err)
	}
	return nil
}
