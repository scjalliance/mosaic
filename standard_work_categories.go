package mosaic

import (
	"context"
	"fmt"
	"strconv"
)

// StandardWorkCategory represents a standard work category in Mosaic.
// Work categories classify the type of work being performed for time tracking.
type StandardWorkCategory struct {
	// MosaicID is the unique identifier for the standard work category in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this standard work category belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Title is the display name of the standard work category.
	Title string `json:"title"`

	// IsBillable indicates whether work in this category is billable.
	IsBillable bool `json:"is_billable,omitempty"`

	// IsArchived indicates whether this standard work category has been archived.
	IsArchived bool `json:"is_archived,omitempty"`

	// ActivityCode is the optional activity code associated with this category.
	ActivityCode *string `json:"activity_code,omitempty"`

	// CreatedAt is the timestamp when the standard work category was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the standard work category was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// standardWorkCategoryListResponse is the wrapper for the GET /api/{team_id}/standard_work_category response,
// which returns {"standard_work_categories": [...]}.
type standardWorkCategoryListResponse struct {
	StandardWorkCategories []StandardWorkCategory `json:"standard_work_categories"`
}

// ListStandardWorkCategories retrieves all standard work categories.
// Uses GET /api/{team_id}/standard_work_category.
func (c *Client) ListStandardWorkCategories(ctx context.Context) ([]StandardWorkCategory, error) {
	var resp standardWorkCategoryListResponse
	path := c.apiPath("standard_work_category")
	if err := c.get(ctx, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing standard work categories: %w", err)
	}
	return resp.StandardWorkCategories, nil
}

// CreateStandardWorkCategory creates a new standard work category and returns the created category.
// Uses POST /api/{team_id}/standard_work_category.
func (c *Client) CreateStandardWorkCategory(ctx context.Context, category *StandardWorkCategory) (*StandardWorkCategory, error) {
	var created StandardWorkCategory
	path := c.apiPath("standard_work_category")
	if err := c.post(ctx, path, category, &created); err != nil {
		return nil, fmt.Errorf("creating standard work category: %w", err)
	}
	return &created, nil
}

// UpdateStandardWorkCategory updates an existing standard work category and returns the updated category.
// Uses PUT /api/{team_id}/standard_work_category/{id}.
func (c *Client) UpdateStandardWorkCategory(ctx context.Context, id int, category *StandardWorkCategory) (*StandardWorkCategory, error) {
	var updated StandardWorkCategory
	path := c.apiPath("standard_work_category", strconv.Itoa(id))
	if err := c.put(ctx, path, category, &updated); err != nil {
		return nil, fmt.Errorf("updating standard work category %d: %w", id, err)
	}
	return &updated, nil
}
