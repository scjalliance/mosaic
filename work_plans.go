package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// HourLock describes how a work plan's hour allocation is locked. The API
// returns it as a string enum.
//
// Breaking change: WorkPlan.HourLock was previously typed *bool, which no
// longer matches the API — work plan responses carry values such as
// "total_hour", so every decode failed. Callers that compared against a
// boolean must switch to the HourLock constants below.
type HourLock string

// Known HourLock values, as observed in live API responses.
const (
	HourLockTotal           HourLock = "total_hour"
	HourLockDaily           HourLock = "daily_hour"
	HourLockWeekly          HourLock = "weekly_hour"
	HourLockPercentCapacity HourLock = "percent_capacity"
)

// DayLock describes how a work plan's day allocation is locked. The API
// returns it as a string enum.
//
// Breaking change: WorkPlan.DayLock was previously typed *bool. See HourLock
// for details.
type DayLock string

// Known DayLock values, as observed in live API responses.
const (
	DayLockNone    DayLock = "none"
	DayLockWorkDay DayLock = "work_day"
)

// WorkPlan represents a work plan entry in Mosaic.
// Field names correspond to the snake_case JSON fields used by the Mosaic API.
type WorkPlan struct {
	// MosaicID is the unique identifier for the work plan.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team this work plan belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// MemberID is the ID of the team member assigned to the work plan.
	MemberID int `json:"member_id,omitempty"`

	// ProjectID is the ID of the associated project.
	ProjectID int `json:"project_id,omitempty"`

	// PhaseID is the ID of the associated phase.
	PhaseID int `json:"phase_id,omitempty"`

	// RoleID is the ID of the role assigned to the work plan.
	RoleID int `json:"role_id,omitempty"`

	// StandardWorkCategoryID is the associated standard work category.
	StandardWorkCategoryID int `json:"standard_work_category_id,omitempty"`

	// Description is a description of the work plan.
	Description string `json:"description,omitempty"`

	// StartDate is the start date of the work plan.
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the end date of the work plan.
	EndDate string `json:"end_date,omitempty"`

	// ActivityPhaseID is the activity phase this work plan is associated with.
	ActivityPhaseID int `json:"activity_phase_id,omitempty"`

	// DailyHours is the number of hours planned per day.
	// Returned as a string by the API (e.g. "0.5").
	DailyHours string `json:"daily_hours,omitempty"`

	// TotalHours is the total number of hours planned.
	// Returned as a string by the API (e.g. "20.0").
	TotalHours string `json:"total_hours,omitempty"`

	// ProjectTitle is the title of the associated project.
	ProjectTitle string `json:"project_title,omitempty"`

	// StandardWorkCategoryTitle is the title of the associated work category.
	StandardWorkCategoryTitle string `json:"standard_work_category_title,omitempty"`

	// MemberName is the display name of the assigned member.
	MemberName string `json:"member_name,omitempty"`

	// MemberEmail is the email of the assigned member.
	MemberEmail string `json:"member_email,omitempty"`

	// PhaseName is the name of the associated phase.
	PhaseName string `json:"phase_name,omitempty"`

	// PortfolioID is the ID of the associated portfolio.
	PortfolioID int `json:"portfolio_id,omitempty"`

	// PortfolioName is the name of the associated portfolio.
	PortfolioName string `json:"portfolio_name,omitempty"`

	// BudgetStatus is the budget status of the work plan.
	BudgetStatus string `json:"budget_status,omitempty"`

	// DayLock describes which day allocation is locked. See the DayLock
	// constants for known values.
	DayLock DayLock `json:"day_lock,omitempty"`

	// HourLock describes how the hour allocation is locked. See the HourLock
	// constants for known values.
	HourLock HourLock `json:"hour_lock,omitempty"`

	// LockHour is the locked hour value when hour lock is enabled.
	LockHour float64 `json:"lock_hour,omitempty"`

	// CreatedAt is the timestamp when the work plan was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the work plan was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// WorkPlanFilter contains filter parameters for listing work plans.
type WorkPlanFilter struct {
	// ListParams contains pagination parameters (Limit, Offset, All).
	ListParams

	// StartDate filters work plans starting on or after this date.
	StartDate string

	// EndDate filters work plans ending on or before this date.
	EndDate string

	// MemberIDs filters by member IDs.
	MemberIDs []int

	// ProjectIDs filters by project IDs.
	ProjectIDs []int

	// PhaseIDs filters by phase IDs.
	PhaseIDs []int

	// StandardWorkCategoryIDs filters by standard work category IDs.
	StandardWorkCategoryIDs []int

	// All fetches all records if set to true.
	All bool

	// MinDailyHours filters work plans with at least this many daily hours.
	MinDailyHours float64

	// MaxDailyHours filters work plans with at most this many daily hours.
	MaxDailyHours float64

	// MinTotalHours filters work plans with at least this many total hours.
	MinTotalHours float64

	// MaxTotalHours filters work plans with at most this many total hours.
	MaxTotalHours float64
}

// toQuery converts WorkPlanFilter into URL query parameter values.
func (f WorkPlanFilter) toQuery() url.Values {
	q := f.ListParams.toQuery()
	if f.StartDate != "" {
		q.Set("start_date", f.StartDate)
	}
	if f.EndDate != "" {
		q.Set("end_date", f.EndDate)
	}
	for _, id := range f.MemberIDs {
		q.Add("member_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.ProjectIDs {
		q.Add("project_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.PhaseIDs {
		q.Add("phase_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.StandardWorkCategoryIDs {
		q.Add("standard_work_category_ids[]", strconv.Itoa(id))
	}
	if f.All {
		q.Set("all", "true")
	}
	if f.MinDailyHours != 0 {
		q.Set("min_daily_hours", strconv.FormatFloat(f.MinDailyHours, 'f', -1, 64))
	}
	if f.MaxDailyHours != 0 {
		q.Set("max_daily_hours", strconv.FormatFloat(f.MaxDailyHours, 'f', -1, 64))
	}
	if f.MinTotalHours != 0 {
		q.Set("min_total_hours", strconv.FormatFloat(f.MinTotalHours, 'f', -1, 64))
	}
	if f.MaxTotalHours != 0 {
		q.Set("max_total_hours", strconv.FormatFloat(f.MaxTotalHours, 'f', -1, 64))
	}
	return q
}

// workPlanBulkDeleteRequest is the request body for bulk deleting work plans.
type workPlanBulkDeleteRequest struct {
	ProjectID int `json:"project_id"`
	PhaseID   int `json:"phase_id"`
}

// workPlanListResponse is the wrapper for the GET /api/{team_id}/work_plan/index response.
type workPlanListResponse struct {
	Count     int        `json:"count"`
	WorkPlans []WorkPlan `json:"work_plans"`
}

// ListWorkPlans retrieves work plans matching the given filter.
// Uses GET /api/{team_id}/work_plan/index.
func (c *Client) ListWorkPlans(ctx context.Context, filter WorkPlanFilter) ([]WorkPlan, error) {
	var resp workPlanListResponse
	path := c.apiPath("work_plan", "index")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing work plans: %w", err)
	}
	return resp.WorkPlans, nil
}

// CreateWorkPlan creates a new work plan and returns the created work plan.
// Uses POST /api/{team_id}/work_plan.
func (c *Client) CreateWorkPlan(ctx context.Context, workPlan *WorkPlan) (*WorkPlan, error) {
	var created WorkPlan
	path := c.apiPath("work_plan")
	if err := c.post(ctx, path, workPlan, &created); err != nil {
		return nil, fmt.Errorf("creating work plan: %w", err)
	}
	return &created, nil
}

// UpdateWorkPlan updates an existing work plan and returns the updated work plan.
// Uses PUT /api/{team_id}/work_plan/{id}.
func (c *Client) UpdateWorkPlan(ctx context.Context, id int, workPlan *WorkPlan) (*WorkPlan, error) {
	var updated WorkPlan
	path := c.apiPath("work_plan", strconv.Itoa(id))
	if err := c.put(ctx, path, workPlan, &updated); err != nil {
		return nil, fmt.Errorf("updating work plan %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteWorkPlan deletes a work plan by ID.
// Uses DELETE /api/{team_id}/work_plan/{id}.
func (c *Client) DeleteWorkPlan(ctx context.Context, id int) error {
	path := c.apiPath("work_plan", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting work plan %d: %w", id, err)
	}
	return nil
}

// BulkDeleteWorkPlans deletes all work plans for a given project and phase.
// Uses DELETE /api/{team_id}/work_plan/bulk.
func (c *Client) BulkDeleteWorkPlans(ctx context.Context, projectID, phaseID int) error {
	path := c.apiPath("work_plan", "bulk")
	body := workPlanBulkDeleteRequest{
		ProjectID: projectID,
		PhaseID:   phaseID,
	}
	if err := c.deleteWithBody(ctx, path, body); err != nil {
		return fmt.Errorf("bulk deleting work plans for project %d phase %d: %w", projectID, phaseID, err)
	}
	return nil
}
