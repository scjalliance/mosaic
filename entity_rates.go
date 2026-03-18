package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// EntityRate represents a rate assignment for an entity in Mosaic.
type EntityRate struct {
	// MosaicID is the unique identifier for the entity rate in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this entity rate belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// EntityID is the ID of the entity this rate applies to.
	EntityID int `json:"entity_id,omitempty"`

	// EntityType is the type of entity this rate applies to.
	EntityType string `json:"entity_type,omitempty"`

	// RateID is the ID of the rate definition.
	RateID int `json:"rate_id,omitempty"`

	// RateGroupID is the ID of the rate group this rate belongs to.
	RateGroupID int `json:"rate_group_id,omitempty"`

	// RateAmount is the monetary rate amount.
	RateAmount float64 `json:"rate_amount,omitempty"`

	// IsCostRate indicates whether this is a cost rate rather than a bill rate.
	IsCostRate *bool `json:"is_cost_rate,omitempty"`

	// OverrideUnassignedMemberRates indicates whether this rate overrides unassigned member rates.
	OverrideUnassignedMemberRates *bool `json:"override_unassigned_member_rates,omitempty"`

	// StartDate is the date the rate assignment begins.
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the date the rate assignment ends.
	EndDate string `json:"end_date,omitempty"`

	// CreatedAt is the timestamp when the entity rate was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the entity rate was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`

	// ArchivedAt is the timestamp when the entity rate was archived, or nil if not archived.
	ArchivedAt *string `json:"archived_at,omitempty"`
}

// EntityRateFilter contains filter parameters for listing entity rates.
type EntityRateFilter struct {
	// EntityType filters entity rates by entity type.
	EntityType string

	// RateGroupIDs filters entity rates by rate group IDs.
	RateGroupIDs []int

	// StartDate filters entity rates starting on or after this date.
	StartDate string

	// EndDate filters entity rates ending on or before this date.
	EndDate string

	// IncludeArchived includes archived entity rates when set to true.
	IncludeArchived *bool
}

// toQuery converts EntityRateFilter into URL query parameter values.
func (f EntityRateFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.EntityType != "" {
		q.Set("entity_type", f.EntityType)
	}
	for _, id := range f.RateGroupIDs {
		q.Add("rate_group_ids[]", strconv.Itoa(id))
	}
	if f.StartDate != "" {
		q.Set("start_date", f.StartDate)
	}
	if f.EndDate != "" {
		q.Set("end_date", f.EndDate)
	}
	if f.IncludeArchived != nil {
		q.Set("include_archived", strconv.FormatBool(*f.IncludeArchived))
	}
	return q
}

// entityRateListResponse is the wrapper for the GET /api/{team_id}/entity_rate response,
// which returns {"entity_rates": [...]}.
type entityRateListResponse struct {
	EntityRates []EntityRate `json:"entity_rates"`
}

// entityRateDeleteRequest is the request body for the DELETE /api/{team_id}/entity_rate/{id} endpoint,
// which accepts an optional rate_group_id.
type entityRateDeleteRequest struct {
	RateGroupID int `json:"rate_group_id,omitempty"`
}

// ListEntityRates retrieves entity rates matching the given filter.
// Uses GET /api/{team_id}/entity_rate.
func (c *Client) ListEntityRates(ctx context.Context, filter EntityRateFilter) ([]EntityRate, error) {
	var resp entityRateListResponse
	path := c.apiPath("entity_rate")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing entity rates: %w", err)
	}
	return resp.EntityRates, nil
}

// CreateEntityRate creates a new entity rate and returns the created entity rate.
// Uses POST /api/{team_id}/entity_rate.
func (c *Client) CreateEntityRate(ctx context.Context, rate *EntityRate) (*EntityRate, error) {
	var created EntityRate
	path := c.apiPath("entity_rate")
	if err := c.post(ctx, path, rate, &created); err != nil {
		return nil, fmt.Errorf("creating entity rate: %w", err)
	}
	return &created, nil
}

// UpdateEntityRate updates an existing entity rate and returns the updated entity rate.
// Uses PUT /api/{team_id}/entity_rate/{entity_rate_id}.
func (c *Client) UpdateEntityRate(ctx context.Context, id int, rate *EntityRate) (*EntityRate, error) {
	var updated EntityRate
	path := c.apiPath("entity_rate", strconv.Itoa(id))
	if err := c.put(ctx, path, rate, &updated); err != nil {
		return nil, fmt.Errorf("updating entity rate %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteEntityRate deletes an entity rate by ID with an optional rate group ID.
// Uses DELETE /api/{team_id}/entity_rate/{entity_rate_id}.
func (c *Client) DeleteEntityRate(ctx context.Context, id int, rateGroupID int) error {
	path := c.apiPath("entity_rate", strconv.Itoa(id))
	body := entityRateDeleteRequest{RateGroupID: rateGroupID}
	if err := c.deleteWithBody(ctx, path, body); err != nil {
		return fmt.Errorf("deleting entity rate %d: %w", id, err)
	}
	return nil
}
