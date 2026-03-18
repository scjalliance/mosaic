package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// PTO represents a paid time off type in Mosaic.
// Field names correspond to the snake_case JSON fields returned by the Mosaic API.
type PTO struct {
	// MosaicID is the unique identifier for the PTO type in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this PTO type belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Name is the display name of the PTO type.
	Name string `json:"name"`

	// Hours is the number of hours for this PTO type.
	Hours float64 `json:"hours,omitempty"`

	// IsAccrued indicates whether this PTO type is accrued over time.
	IsAccrued *bool `json:"is_accrued,omitempty"`

	// IsCustom indicates whether this is a custom PTO type.
	IsCustom *bool `json:"is_custom,omitempty"`

	// CreatedAt is the timestamp when the PTO type was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the PTO type was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// PTOFilter contains filter parameters for listing PTO types.
type PTOFilter struct {
	// IncludeArchived includes archived PTO types when set to true.
	IncludeArchived *bool

	// IncludeDefault includes default PTO types when set to true.
	IncludeDefault *bool

	// IsCustom filters PTO types by custom status.
	IsCustom *bool
}

// toQuery converts PTOFilter into URL query parameter values.
func (f PTOFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.IncludeArchived != nil {
		q.Set("include_archived", strconv.FormatBool(*f.IncludeArchived))
	}
	if f.IncludeDefault != nil {
		q.Set("include_default", strconv.FormatBool(*f.IncludeDefault))
	}
	if f.IsCustom != nil {
		q.Set("is_custom", strconv.FormatBool(*f.IsCustom))
	}
	return q
}

// ptoListResponse is the wrapper for the GET /api/{team_id}/pto response,
// which returns {"pto_policies": [...]}.
type ptoListResponse struct {
	PTOPolicies []PTO `json:"pto_policies"`
}

// ListPTOs retrieves PTO types matching the given filter.
// Uses GET /api/{team_id}/pto.
func (c *Client) ListPTOs(ctx context.Context, filter PTOFilter) ([]PTO, error) {
	var resp ptoListResponse
	path := c.apiPath("pto")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing PTOs: %w", err)
	}
	return resp.PTOPolicies, nil
}

// CreatePTO creates a new PTO type and returns the created PTO.
// Uses POST /api/{team_id}/pto.
func (c *Client) CreatePTO(ctx context.Context, pto *PTO) (*PTO, error) {
	var created PTO
	path := c.apiPath("pto")
	if err := c.post(ctx, path, pto, &created); err != nil {
		return nil, fmt.Errorf("creating PTO: %w", err)
	}
	return &created, nil
}

// UpdatePTO updates an existing PTO type and returns the updated PTO.
// Uses PUT /api/{team_id}/pto/{pto_id}.
func (c *Client) UpdatePTO(ctx context.Context, id int, pto *PTO) (*PTO, error) {
	var updated PTO
	path := c.apiPath("pto", strconv.Itoa(id))
	if err := c.put(ctx, path, pto, &updated); err != nil {
		return nil, fmt.Errorf("updating PTO %d: %w", id, err)
	}
	return &updated, nil
}

// DeletePTO deletes a PTO type by ID.
// Uses DELETE /api/{team_id}/pto/{pto_id}.
func (c *Client) DeletePTO(ctx context.Context, id int) error {
	path := c.apiPath("pto", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting PTO %d: %w", id, err)
	}
	return nil
}
