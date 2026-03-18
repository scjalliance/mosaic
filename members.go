package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Member represents a team member (employee) in Mosaic.
// Field names correspond to the snake_case JSON fields returned by the Mosaic API.
// The API resource is called "employee" but exposed here as "Member" for clarity.
type Member struct {
	// MosaicID is the unique identifier for the member in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this member belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// FirstName is the member's first name.
	FirstName string `json:"first_name,omitempty"`

	// LastName is the member's last name.
	LastName string `json:"last_name,omitempty"`

	// Email is the member's email address.
	Email string `json:"email,omitempty"`

	// EmploymentType is the member's employment type.
	EmploymentType string `json:"employment_type,omitempty"`

	// IsArchived indicates whether the member is archived.
	IsArchived bool `json:"is_archived,omitempty"`

	// OfficeIDs is the list of office IDs the member is associated with.
	OfficeIDs []int `json:"office_ids,omitempty"`

	// Offices is the list of office names the member is associated with.
	Offices []string `json:"offices,omitempty"`

	// Regions is the list of region names the member is associated with.
	Regions []string `json:"regions,omitempty"`

	// WorkGroups is the list of work group IDs the member belongs to.
	WorkGroups []int `json:"work_groups,omitempty"`

	// Locations is the member's location information. Each location is an
	// arbitrary object returned by the API, represented as a map.
	Locations []map[string]any `json:"locations,omitempty"`

	// CreatedAt is the timestamp when the member was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the member was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// memberListResponse is the wrapper struct for the list members API response.
// The API returns {"employee": [...]}.
type memberListResponse struct {
	Employee []Member `json:"employee"`
}

// ListAllMembers retrieves all members (employees) in the team.
// Uses GET /api/{team_id}/employee/index.
func (c *Client) ListAllMembers(ctx context.Context, includeDiscarded bool) ([]Member, error) {
	var resp memberListResponse
	path := c.apiPath("employee", "index")
	q := make(url.Values)
	if includeDiscarded {
		q.Set("include_discarded", "true")
	}
	if err := c.get(ctx, path, q, &resp); err != nil {
		return nil, fmt.Errorf("listing all members: %w", err)
	}
	return resp.Employee, nil
}

// GetMember retrieves a single member by their member ID.
// Uses GET /api/{team_id}/employee?member_id=X.
func (c *Client) GetMember(ctx context.Context, memberID int) (*Member, error) {
	var member Member
	path := c.apiPath("employee")
	q := make(url.Values)
	q.Set("member_id", strconv.Itoa(memberID))
	if err := c.get(ctx, path, q, &member); err != nil {
		return nil, fmt.Errorf("getting member %d: %w", memberID, err)
	}
	return &member, nil
}

// CreateMember creates a new member (employee) and returns the created member.
// Uses POST /api/{team_id}/employee.
func (c *Client) CreateMember(ctx context.Context, member *Member) (*Member, error) {
	var created Member
	path := c.apiPath("employee")
	if err := c.post(ctx, path, member, &created); err != nil {
		return nil, fmt.Errorf("creating member: %w", err)
	}
	return &created, nil
}

// UpdateMember updates an existing member and returns the updated member.
// Uses PUT /api/{team_id}/employee (body, not path param).
func (c *Client) UpdateMember(ctx context.Context, member *Member) (*Member, error) {
	var updated Member
	path := c.apiPath("employee")
	if err := c.put(ctx, path, member, &updated); err != nil {
		return nil, fmt.Errorf("updating member: %w", err)
	}
	return &updated, nil
}

// DeleteMember deletes a member (employee).
// Uses DELETE /api/{team_id}/employee with a request body.
func (c *Client) DeleteMember(ctx context.Context) error {
	path := c.apiPath("employee")
	if err := c.deleteWithBody(ctx, path, nil); err != nil {
		return fmt.Errorf("deleting member: %w", err)
	}
	return nil
}
