package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Scope represents a project scope (work assignment) in Mosaic.
// Scopes are units of work within activity phases that can be assigned to members.
type Scope struct {
	// MosaicID is the unique identifier for the scope.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team this scope belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// ActivityPhaseID is the activity phase this scope belongs to.
	ActivityPhaseID int `json:"activity_phase_id,omitempty"`

	// ActivityID is the activity this scope is associated with.
	ActivityID int `json:"activity_id,omitempty"`

	// Description is the scope description.
	Description string `json:"description,omitempty"`

	// EstimatedHours is the estimated hours for this scope.
	EstimatedHours float64 `json:"estimated_hours,omitempty"`

	// IsRequest indicates whether this scope is a request.
	IsRequest *bool `json:"is_request,omitempty"`

	// Note is an additional note on the scope.
	Note string `json:"note,omitempty"`

	// ParentScopeID is the parent scope ID for nested scopes.
	ParentScopeID int `json:"parent_scope_id,omitempty"`

	// RequestedToID is the member ID this scope is requested to.
	RequestedToID int `json:"requested_to_id,omitempty"`

	// ScheduleStart is the scheduled start date.
	ScheduleStart string `json:"schedule_start,omitempty"`

	// ScheduleEnd is the scheduled end date.
	ScheduleEnd string `json:"schedule_end,omitempty"`

	// CreatedAt is the timestamp when the scope was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the scope was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ScopeFilter contains filter parameters for listing scopes.
type ScopeFilter struct {
	ListParams

	// ProjectIDs filters scopes by project IDs.
	ProjectIDs []int

	// PhaseIDs filters scopes by phase IDs.
	PhaseIDs []int

	// ActivityPhaseIDs filters scopes by activity phase IDs.
	ActivityPhaseIDs []int

	// MemberIDs filters scopes by assigned member IDs.
	MemberIDs []int

	// ScopeIDs filters to specific scope IDs.
	ScopeIDs []int

	// ScheduleStart filters scopes starting on or after this date.
	ScheduleStart string

	// ScheduleEnd filters scopes ending on or before this date.
	ScheduleEnd string

	// All returns all scopes when true.
	All bool

	// IncludeRequests includes request scopes.
	IncludeRequests *bool

	// ExcludeNonRequests excludes non-request scopes.
	ExcludeNonRequests *bool

	// GroupDeterminant controls grouping of results.
	GroupDeterminant string
}

// toQuery converts ScopeFilter into URL query parameter values.
func (f ScopeFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.Limit > 0 {
		q.Set("limit", strconv.Itoa(f.Limit))
	}
	if f.Offset > 0 {
		q.Set("offset", strconv.Itoa(f.Offset))
	}
	for _, id := range f.ProjectIDs {
		q.Add("project_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.PhaseIDs {
		q.Add("phase_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.ActivityPhaseIDs {
		q.Add("activity_phase_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.MemberIDs {
		q.Add("member_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.ScopeIDs {
		q.Add("scope_ids[]", strconv.Itoa(id))
	}
	if f.ScheduleStart != "" {
		q.Set("schedule_start", f.ScheduleStart)
	}
	if f.ScheduleEnd != "" {
		q.Set("schedule_end", f.ScheduleEnd)
	}
	if f.All {
		q.Set("all", "true")
	}
	if f.IncludeRequests != nil {
		q.Set("include_requests", strconv.FormatBool(*f.IncludeRequests))
	}
	if f.ExcludeNonRequests != nil {
		q.Set("exclude_non_requests", strconv.FormatBool(*f.ExcludeNonRequests))
	}
	if f.GroupDeterminant != "" {
		q.Set("group_determinant", f.GroupDeterminant)
	}
	return q
}

// scopeListResponse is the wrapper for the GET /api/{team_id}/scope response.
type scopeListResponse struct {
	Scopes []Scope `json:"scopes"`
}

// ListScopes retrieves scopes matching the given filter.
// Uses GET /api/{team_id}/scope.
func (c *Client) ListScopes(ctx context.Context, filter ScopeFilter) ([]Scope, error) {
	var resp scopeListResponse
	path := c.apiPath("scope")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing scopes: %w", err)
	}
	return resp.Scopes, nil
}

// CreateScope creates a new project scope.
// Uses POST /api/{team_id}/scope.
func (c *Client) CreateScope(ctx context.Context, scope *Scope) (*Scope, error) {
	var created Scope
	path := c.apiPath("scope")
	if err := c.post(ctx, path, scope, &created); err != nil {
		return nil, fmt.Errorf("creating scope: %w", err)
	}
	return &created, nil
}

// UpdateScope updates an existing project scope.
// Uses PUT /api/{team_id}/scope/{scope_id}.
func (c *Client) UpdateScope(ctx context.Context, scopeID int, scope *Scope) (*Scope, error) {
	var updated Scope
	path := c.apiPath("scope", strconv.Itoa(scopeID))
	if err := c.put(ctx, path, scope, &updated); err != nil {
		return nil, fmt.Errorf("updating scope %d: %w", scopeID, err)
	}
	return &updated, nil
}

// DeleteScope deletes a project scope.
// Uses DELETE /api/{team_id}/scope/{scope_id}.
func (c *Client) DeleteScope(ctx context.Context, scopeID int) error {
	path := c.apiPath("scope", strconv.Itoa(scopeID))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting scope %d: %w", scopeID, err)
	}
	return nil
}

// scopeAssignRequest is the request body for assigning members to a scope.
type scopeAssignRequest struct {
	Members []int `json:"members"`
}

// AssignScopeMembers assigns members to a project scope.
// Uses POST /api/{team_id}/scope/{scope_id}/assign.
func (c *Client) AssignScopeMembers(ctx context.Context, scopeID int, memberIDs []int) error {
	path := c.apiPath("scope", strconv.Itoa(scopeID), "assign")
	body := scopeAssignRequest{Members: memberIDs}
	if err := c.post(ctx, path, body, nil); err != nil {
		return fmt.Errorf("assigning members to scope %d: %w", scopeID, err)
	}
	return nil
}

// ScopeAssignmentUpdate contains fields for updating a scope assignment.
type ScopeAssignmentUpdate struct {
	// StartDate is the assignment start date.
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the assignment end date.
	EndDate string `json:"end_date,omitempty"`

	// PercentComplete is the assignment completion percentage.
	PercentComplete float64 `json:"percent_complete,omitempty"`
}

// UpdateScopeAssignment updates a member's assignment on a scope.
// Uses PUT /api/{team_id}/scope/{scope_id}/update_assignment/{member_id}.
func (c *Client) UpdateScopeAssignment(ctx context.Context, scopeID, memberID int, update *ScopeAssignmentUpdate) error {
	path := c.apiPath("scope", strconv.Itoa(scopeID), "update_assignment", strconv.Itoa(memberID))
	if err := c.put(ctx, path, update, nil); err != nil {
		return fmt.Errorf("updating scope %d assignment for member %d: %w", scopeID, memberID, err)
	}
	return nil
}
