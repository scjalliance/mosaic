package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ActivityPhase represents an activity phase in Mosaic.
// Field names correspond to the snake_case JSON fields used by the Mosaic API.
type ActivityPhase struct {
	// MosaicID is the unique identifier for the activity phase.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team this activity phase belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// PhaseID is the ID of the parent phase.
	PhaseID int `json:"phase_id,omitempty"`

	// Title is the display name of the activity phase.
	Title string `json:"title"`

	// Billable indicates whether the activity phase is billable.
	Billable *bool `json:"billable,omitempty"`

	// IsCustom indicates whether this is a custom activity phase.
	IsCustom *bool `json:"is_custom,omitempty"`

	// IsArchived indicates whether the activity phase has been archived.
	IsArchived *bool `json:"is_archived,omitempty"`

	// StandardWorkCategoryID is the associated standard work category.
	StandardWorkCategoryID int `json:"standard_work_category_id,omitempty"`

	// StandardWorkCategoryTitle is the title of the associated standard work category.
	StandardWorkCategoryTitle string `json:"standard_work_category_title,omitempty"`

	// ProjectID is the project this activity phase belongs to.
	ProjectID int `json:"project_id,omitempty"`

	// IsDefault indicates whether this is a default activity phase.
	IsDefault bool `json:"is_default,omitempty"`

	// StartDate is the start date of the activity phase.
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the end date of the activity phase.
	EndDate string `json:"end_date,omitempty"`

	// EstimatedHours is the estimated number of hours for the activity phase.
	EstimatedHours float64 `json:"estimated_hours,omitempty"`

	// EstimatedCost is the estimated cost for the activity phase.
	EstimatedCost float64 `json:"estimated_cost,omitempty"`

	// Total is the total value of the activity phase.
	// Returned as a string by the API (e.g. "0.0").
	Total string `json:"total,omitempty"`

	// FeeType is the fee type (e.g. "fixed", "hourly").
	FeeType string `json:"fee_type,omitempty"`

	// RateMultiplier is the multiplier applied to the billing rate.
	// Returned as a string by the API (e.g. "1.0").
	RateMultiplier string `json:"rate_multiplier,omitempty"`

	// CostRateMultiplier is the multiplier applied to the cost rate.
	// Returned as a string by the API (e.g. "1.0").
	CostRateMultiplier string `json:"cost_rate_multiplier,omitempty"`

	// ProfitPercentage is the target profit percentage.
	ProfitPercentage float64 `json:"profit_percentage,omitempty"`

	// PercentageOfPhase is the percentage of the parent phase this activity represents.
	PercentageOfPhase float64 `json:"percentage_of_phase,omitempty"`

	// BudgetActivityPhaseBy controls how the activity phase budget is calculated.
	BudgetActivityPhaseBy string `json:"budget_activity_phase_by,omitempty"`

	// BudgetFixedFeeWith controls what is used for fixed fee budget calculation.
	BudgetFixedFeeWith string `json:"budget_fixed_fee_with,omitempty"`

	// BudgetHourlyWith controls what is used for hourly budget calculation.
	BudgetHourlyWith string `json:"budget_hourly_with,omitempty"`

	// BudgetInternalWith controls what is used for internal budget calculation.
	BudgetInternalWith string `json:"budget_internal_with,omitempty"`

	// CreatedAt is the timestamp when the activity phase was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the activity phase was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ActivityPhaseFilter contains filter parameters for listing activity phases.
type ActivityPhaseFilter struct {
	// ProjectIDs filters activity phases by project IDs.
	ProjectIDs []int

	// PhaseIDs filters activity phases by phase IDs.
	PhaseIDs []int

	// StartDate filters activity phases starting on or after this date.
	StartDate string

	// EndDate filters activity phases ending on or before this date.
	EndDate string

	// All fetches all records if set to true.
	All bool
}

// toQuery converts ActivityPhaseFilter into URL query parameter values.
func (f ActivityPhaseFilter) toQuery() url.Values {
	q := make(url.Values)
	for _, id := range f.ProjectIDs {
		q.Add("project_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.PhaseIDs {
		q.Add("phase_ids[]", strconv.Itoa(id))
	}
	if f.StartDate != "" {
		q.Set("start_date", f.StartDate)
	}
	if f.EndDate != "" {
		q.Set("end_date", f.EndDate)
	}
	if f.All {
		q.Set("all", "true")
	}
	return q
}

// activityPhaseListResponse is the wrapper for the GET /api/{team_id}/activity_phase response.
// The API returns activity phases under the "work_categories" key.
type activityPhaseListResponse struct {
	WorkCategories []ActivityPhase `json:"work_categories"`
}

// activityPhaseDeleteRequest is the request body for deleting an activity phase.
type activityPhaseDeleteRequest struct {
	ForceDestroy bool `json:"force_destroy,omitempty"`
}

// activityPhaseConvertRequest is the request body for converting an activity phase to a subphase.
type activityPhaseConvertRequest struct {
	ActivityPhaseID int  `json:"activity_phase_id"`
	KeepActivity    bool `json:"keep_activity"`
}

// ListActivityPhases retrieves activity phases matching the given filter.
// Uses GET /api/{team_id}/activity_phase.
func (c *Client) ListActivityPhases(ctx context.Context, filter ActivityPhaseFilter) ([]ActivityPhase, error) {
	var resp activityPhaseListResponse
	path := c.apiPath("activity_phase")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing activity phases: %w", err)
	}
	return resp.WorkCategories, nil
}

// CreateActivityPhase creates a new activity phase and returns the created activity phase.
// Uses POST /api/{team_id}/activity_phase.
func (c *Client) CreateActivityPhase(ctx context.Context, activityPhase *ActivityPhase) (*ActivityPhase, error) {
	var created ActivityPhase
	path := c.apiPath("activity_phase")
	if err := c.post(ctx, path, activityPhase, &created); err != nil {
		return nil, fmt.Errorf("creating activity phase: %w", err)
	}
	return &created, nil
}

// UpdateActivityPhase updates an existing activity phase and returns the updated activity phase.
// Uses PUT /api/{team_id}/activity_phase/{activity_phase_id}.
func (c *Client) UpdateActivityPhase(ctx context.Context, id int, activityPhase *ActivityPhase) (*ActivityPhase, error) {
	var updated ActivityPhase
	path := c.apiPath("activity_phase", strconv.Itoa(id))
	if err := c.put(ctx, path, activityPhase, &updated); err != nil {
		return nil, fmt.Errorf("updating activity phase %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteActivityPhase deletes an activity phase by ID. If forceDestroy is true, the activity
// phase will be deleted even if it has associated data.
// Uses DELETE /api/{team_id}/activity_phase/{activity_phase_id}.
func (c *Client) DeleteActivityPhase(ctx context.Context, id int, forceDestroy bool) error {
	path := c.apiPath("activity_phase", strconv.Itoa(id))
	body := activityPhaseDeleteRequest{ForceDestroy: forceDestroy}
	if err := c.deleteWithBody(ctx, path, body); err != nil {
		return fmt.Errorf("deleting activity phase %d: %w", id, err)
	}
	return nil
}

// ConvertActivityPhaseToSubphase converts an activity phase to a subphase.
// If keepActivity is true, the original activity is preserved during conversion.
// Uses POST /api/{team_id}/activity_phase/convert_to_subphase.
func (c *Client) ConvertActivityPhaseToSubphase(ctx context.Context, activityPhaseID int, keepActivity bool) error {
	path := c.apiPath("activity_phase", "convert_to_subphase")
	body := activityPhaseConvertRequest{
		ActivityPhaseID: activityPhaseID,
		KeepActivity:    keepActivity,
	}
	if err := c.post(ctx, path, body, nil); err != nil {
		return fmt.Errorf("converting activity phase %d to subphase: %w", activityPhaseID, err)
	}
	return nil
}
