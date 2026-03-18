package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// MemberRole represents a role assignment for a member in Mosaic.
type MemberRole struct {
	// MosaicID is the unique identifier for the member role in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this member role belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// MemberID is the ID of the member assigned the role.
	MemberID int `json:"member_id,omitempty"`

	// RoleID is the ID of the role assigned to the member.
	RoleID int `json:"role_id,omitempty"`

	// StartDate is the date the role assignment begins.
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the date the role assignment ends.
	EndDate string `json:"end_date,omitempty"`

	// OverrideMemberPositions indicates whether this role overrides the member's default positions.
	OverrideMemberPositions *bool `json:"override_member_positions,omitempty"`

	// CreatedAt is the timestamp when the member role was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the member role was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`

	// TeamMembershipID is the team membership ID associated with this role.
	TeamMembershipID int `json:"team_membership_id,omitempty"`
}

// MemberRoleFilter contains filter parameters for listing member roles.
type MemberRoleFilter struct {
	// MemberID filters member roles by member ID.
	MemberID int

	// RoleID filters member roles by role ID.
	RoleID int

	// IncludeArchived includes archived member roles when set to true.
	IncludeArchived *bool

	// IsActive filters member roles by active status.
	IsActive *bool
}

// toQuery converts MemberRoleFilter into URL query parameter values.
func (f MemberRoleFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.MemberID != 0 {
		q.Set("member_id", strconv.Itoa(f.MemberID))
	}
	if f.RoleID != 0 {
		q.Set("role_id", strconv.Itoa(f.RoleID))
	}
	if f.IncludeArchived != nil {
		q.Set("include_archived", strconv.FormatBool(*f.IncludeArchived))
	}
	if f.IsActive != nil {
		q.Set("is_active", strconv.FormatBool(*f.IsActive))
	}
	return q
}

// memberRoleListResponse is the wrapper for the GET /api/{team_id}/member_role response,
// which returns {"team_positions": [...]}.
type memberRoleListResponse struct {
	TeamPositions []MemberRole `json:"team_positions"`
}

// ListMemberRoles retrieves member roles matching the given filter.
// Uses GET /api/{team_id}/member_role.
func (c *Client) ListMemberRoles(ctx context.Context, filter MemberRoleFilter) ([]MemberRole, error) {
	var resp memberRoleListResponse
	path := c.apiPath("member_role")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing member roles: %w", err)
	}
	return resp.TeamPositions, nil
}

// CreateMemberRole creates a new member role and returns the created member role.
// Uses POST /api/{team_id}/member_role.
func (c *Client) CreateMemberRole(ctx context.Context, memberRole *MemberRole) (*MemberRole, error) {
	var created MemberRole
	path := c.apiPath("member_role")
	if err := c.post(ctx, path, memberRole, &created); err != nil {
		return nil, fmt.Errorf("creating member role: %w", err)
	}
	return &created, nil
}

// UpdateMemberRole updates an existing member role and returns the updated member role.
// Uses PUT /api/{team_id}/member_role/{member_role_id}.
func (c *Client) UpdateMemberRole(ctx context.Context, id int, memberRole *MemberRole) (*MemberRole, error) {
	var updated MemberRole
	path := c.apiPath("member_role", strconv.Itoa(id))
	if err := c.put(ctx, path, memberRole, &updated); err != nil {
		return nil, fmt.Errorf("updating member role %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteMemberRole deletes a member role by ID.
// Uses DELETE /api/{team_id}/member_role/{member_role_id}.
func (c *Client) DeleteMemberRole(ctx context.Context, id int) error {
	path := c.apiPath("member_role", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting member role %d: %w", id, err)
	}
	return nil
}
