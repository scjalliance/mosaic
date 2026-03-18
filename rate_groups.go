package mosaic

import (
	"context"
	"fmt"
	"strconv"
)

// RateGroup represents a rate group in Mosaic.
// Field names correspond to the snake_case JSON fields returned by the Mosaic API.
type RateGroup struct {
	// MosaicID is the unique identifier for the rate group in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this rate group belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Name is the display name of the rate group.
	Name string `json:"name"`

	// Description is the rate group description.
	Description string `json:"description,omitempty"`

	// ActiveEntityType is the active entity type for this rate group.
	ActiveEntityType string `json:"active_entity_type,omitempty"`

	// Currency is the currency name for this rate group.
	Currency string `json:"currency,omitempty"`

	// CurrencyCode is the currency code for this rate group (e.g. "USD").
	CurrencyCode string `json:"currency_code,omitempty"`

	// IsDefault indicates whether this is the default rate group.
	IsDefault *bool `json:"is_default,omitempty"`

	// CreatedAt is the timestamp when the rate group was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the rate group was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// rateGroupListResponse is the wrapper for the GET /api/{team_id}/rate_group response,
// which returns {"rate_groups": [...]}.
type rateGroupListResponse struct {
	RateGroups []RateGroup `json:"rate_groups"`
}

// rateGroupDeleteRequest is the request body for deleting a rate group.
type rateGroupDeleteRequest struct {
	ForceDestroy bool `json:"force_destroy,omitempty"`
}

// ListRateGroups retrieves all rate groups.
// Uses GET /api/{team_id}/rate_group.
func (c *Client) ListRateGroups(ctx context.Context) ([]RateGroup, error) {
	var resp rateGroupListResponse
	path := c.apiPath("rate_group")
	if err := c.get(ctx, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing rate groups: %w", err)
	}
	return resp.RateGroups, nil
}

// CreateRateGroup creates a new rate group and returns the created rate group.
// Uses POST /api/{team_id}/rate_group.
func (c *Client) CreateRateGroup(ctx context.Context, rateGroup *RateGroup) (*RateGroup, error) {
	var created RateGroup
	path := c.apiPath("rate_group")
	if err := c.post(ctx, path, rateGroup, &created); err != nil {
		return nil, fmt.Errorf("creating rate group: %w", err)
	}
	return &created, nil
}

// UpdateRateGroup updates an existing rate group and returns the updated rate group.
// Uses PUT /api/{team_id}/rate_group/{rate_group_id}.
func (c *Client) UpdateRateGroup(ctx context.Context, id int, rateGroup *RateGroup) (*RateGroup, error) {
	var updated RateGroup
	path := c.apiPath("rate_group", strconv.Itoa(id))
	if err := c.put(ctx, path, rateGroup, &updated); err != nil {
		return nil, fmt.Errorf("updating rate group %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteRateGroup deletes a rate group by ID with an optional force destroy flag.
// Uses DELETE /api/{team_id}/rate_group/{rate_group_id}.
func (c *Client) DeleteRateGroup(ctx context.Context, id int, forceDestroy bool) error {
	path := c.apiPath("rate_group", strconv.Itoa(id))
	body := rateGroupDeleteRequest{
		ForceDestroy: forceDestroy,
	}
	if err := c.deleteWithBody(ctx, path, body); err != nil {
		return fmt.Errorf("deleting rate group %d: %w", id, err)
	}
	return nil
}
