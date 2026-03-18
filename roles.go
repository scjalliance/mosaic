package mosaic

import (
	"context"
	"fmt"
	"strconv"
)

// Role represents a role within an organization in Mosaic.
type Role struct {
	// MosaicID is the unique identifier for the role in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this role belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Name is the display name of the role.
	Name string `json:"name"`

	// Description is the role's description.
	Description string `json:"description,omitempty"`

	// CreatedAt is the timestamp when the role was created.
	CreatedAt string `json:"created_at,omitempty"`

	// IsDefault indicates whether this is the default role.
	IsDefault bool `json:"is_default,omitempty"`

	// UpdatedAt is the timestamp when the role was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// roleListResponse is the wrapper for the GET /api/{team_id}/role response,
// which returns {"roles": [...]}.
type roleListResponse struct {
	Roles []Role `json:"roles"`
}

// ListRoles retrieves all roles.
// Uses GET /api/{team_id}/role.
func (c *Client) ListRoles(ctx context.Context) ([]Role, error) {
	var resp roleListResponse
	path := c.apiPath("role")
	if err := c.get(ctx, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing roles: %w", err)
	}
	return resp.Roles, nil
}

// CreateRole creates a new role and returns the created role.
// Uses POST /api/{team_id}/role.
func (c *Client) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	var created Role
	path := c.apiPath("role")
	if err := c.post(ctx, path, role, &created); err != nil {
		return nil, fmt.Errorf("creating role: %w", err)
	}
	return &created, nil
}

// UpdateRole updates an existing role and returns the updated role.
// Uses PUT /api/{team_id}/role/{id}.
func (c *Client) UpdateRole(ctx context.Context, id int, role *Role) (*Role, error) {
	var updated Role
	path := c.apiPath("role", strconv.Itoa(id))
	if err := c.put(ctx, path, role, &updated); err != nil {
		return nil, fmt.Errorf("updating role %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteRole deletes a role by ID.
// Uses DELETE /api/{team_id}/role/{id}.
func (c *Client) DeleteRole(ctx context.Context, id int) error {
	path := c.apiPath("role", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting role %d: %w", id, err)
	}
	return nil
}
