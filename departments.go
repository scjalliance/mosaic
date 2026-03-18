package mosaic

import (
	"context"
	"fmt"
	"strconv"
)

// Department represents a department within an organization in Mosaic.
type Department struct {
	// MosaicID is the unique identifier for the department in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this department belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Name is the display name of the department.
	Name string `json:"name"`

	// Description is the department's description.
	Description string `json:"description,omitempty"`

	// CreatedAt is the timestamp when the department was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the department was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// departmentListResponse is the wrapper for the GET /api/{team_id}/department response,
// which returns {"departments": [...]}.
type departmentListResponse struct {
	Departments []Department `json:"departments"`
}

// ListDepartments retrieves all departments.
// Uses GET /api/{team_id}/department.
func (c *Client) ListDepartments(ctx context.Context) ([]Department, error) {
	var resp departmentListResponse
	path := c.apiPath("department")
	if err := c.get(ctx, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing departments: %w", err)
	}
	return resp.Departments, nil
}

// CreateDepartment creates a new department and returns the created department.
// Uses POST /api/{team_id}/department.
func (c *Client) CreateDepartment(ctx context.Context, department *Department) (*Department, error) {
	var created Department
	path := c.apiPath("department")
	if err := c.post(ctx, path, department, &created); err != nil {
		return nil, fmt.Errorf("creating department: %w", err)
	}
	return &created, nil
}

// UpdateDepartment updates an existing department and returns the updated department.
// Uses PUT /api/{team_id}/department/{id}.
func (c *Client) UpdateDepartment(ctx context.Context, id int, department *Department) (*Department, error) {
	var updated Department
	path := c.apiPath("department", strconv.Itoa(id))
	if err := c.put(ctx, path, department, &updated); err != nil {
		return nil, fmt.Errorf("updating department %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteDepartment deletes a department by ID.
// Uses DELETE /api/{team_id}/department/{id}.
func (c *Client) DeleteDepartment(ctx context.Context, id int) error {
	path := c.apiPath("department", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting department %d: %w", id, err)
	}
	return nil
}
