package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// MemberProjectRole represents a role assignment for a member on a specific project in Mosaic.
type MemberProjectRole struct {
	// MosaicID is the unique identifier for the member project role in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this member project role belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// MemberID is the ID of the member assigned the role.
	MemberID int `json:"member_id,omitempty"`

	// ProjectID is the ID of the project this role applies to.
	ProjectID int `json:"project_id,omitempty"`

	// PhaseID is the ID of the phase this role applies to.
	PhaseID int `json:"phase_id,omitempty"`

	// RoleID is the ID of the role assigned to the member on this project.
	RoleID int `json:"role_id,omitempty"`

	// StartDate is the date the role assignment begins.
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the date the role assignment ends.
	EndDate string `json:"end_date,omitempty"`

	// OverridePhaseMemberPositions indicates whether this role overrides the phase member's default positions.
	OverridePhaseMemberPositions *bool `json:"override_phase_member_positions,omitempty"`

	// CreatedAt is the timestamp when the member project role was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the member project role was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`

	// MemberBudgetID is the member budget ID associated with this project role.
	MemberBudgetID int `json:"member_budget_id,omitempty"`
}

// MemberProjectRoleFilter contains filter parameters for listing member project roles.
type MemberProjectRoleFilter struct {
	// MemberID filters member project roles by member ID.
	MemberID int

	// ProjectID filters member project roles by project ID.
	ProjectID int

	// RoleID filters member project roles by role ID.
	RoleID int

	// PhaseID filters member project roles by phase ID.
	PhaseID int

	// IncludeArchived includes archived member project roles when set to true.
	IncludeArchived *bool

	// IsActive filters member project roles by active status.
	IsActive *bool
}

// toQuery converts MemberProjectRoleFilter into URL query parameter values.
func (f MemberProjectRoleFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.MemberID != 0 {
		q.Set("member_id", strconv.Itoa(f.MemberID))
	}
	if f.ProjectID != 0 {
		q.Set("project_id", strconv.Itoa(f.ProjectID))
	}
	if f.RoleID != 0 {
		q.Set("role_id", strconv.Itoa(f.RoleID))
	}
	if f.PhaseID != 0 {
		q.Set("phase_id", strconv.Itoa(f.PhaseID))
	}
	if f.IncludeArchived != nil {
		q.Set("include_archived", strconv.FormatBool(*f.IncludeArchived))
	}
	if f.IsActive != nil {
		q.Set("is_active", strconv.FormatBool(*f.IsActive))
	}
	return q
}

// memberProjectRoleListResponse is the wrapper for the GET /api/{team_id}/member_project_role response,
// which returns {"member_project_roles": [...]}.
type memberProjectRoleListResponse struct {
	MemberProjectRoles []MemberProjectRole `json:"member_project_roles"`
}

// ListMemberProjectRoles retrieves member project roles matching the given filter.
// Uses GET /api/{team_id}/member_project_role.
func (c *Client) ListMemberProjectRoles(ctx context.Context, filter MemberProjectRoleFilter) ([]MemberProjectRole, error) {
	var resp memberProjectRoleListResponse
	path := c.apiPath("member_project_role")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing member project roles: %w", err)
	}
	return resp.MemberProjectRoles, nil
}

// CreateMemberProjectRole creates a new member project role and returns the created member project role.
// Uses POST /api/{team_id}/member_project_role.
func (c *Client) CreateMemberProjectRole(ctx context.Context, role *MemberProjectRole) (*MemberProjectRole, error) {
	var created MemberProjectRole
	path := c.apiPath("member_project_role")
	if err := c.post(ctx, path, role, &created); err != nil {
		return nil, fmt.Errorf("creating member project role: %w", err)
	}
	return &created, nil
}

// UpdateMemberProjectRole updates an existing member project role and returns the updated member project role.
// Uses PUT /api/{team_id}/member_project_role/{member_project_role_id}.
// Note: there is no DELETE endpoint for member project roles.
func (c *Client) UpdateMemberProjectRole(ctx context.Context, id int, role *MemberProjectRole) (*MemberProjectRole, error) {
	var updated MemberProjectRole
	path := c.apiPath("member_project_role", strconv.Itoa(id))
	if err := c.put(ctx, path, role, &updated); err != nil {
		return nil, fmt.Errorf("updating member project role %d: %w", id, err)
	}
	return &updated, nil
}
