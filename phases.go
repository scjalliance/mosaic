package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Phase represents a phase within a project in Mosaic.
// Field names correspond to the snake_case JSON fields returned by the Mosaic API.
type Phase struct {
	// MosaicID is the unique identifier for the phase (API field: mosaic_id).
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this phase belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Name is the display name of the phase.
	Name string `json:"name"`

	// Title is the title of the phase (typically the same as Name).
	Title string `json:"title,omitempty"`

	// ProjectID is the ID of the project this phase belongs to.
	ProjectID int `json:"project_id"`

	// PhaseNumber is the phase number (nullable string).
	PhaseNumber *string `json:"phase_number,omitempty"`

	// BudgetStatus is the current budget status of the phase.
	BudgetStatus string `json:"budget_status,omitempty"`

	// Total is the total budget for the phase (nullable string, e.g. "18818.0").
	Total *string `json:"total,omitempty"`

	// EstimatedCost is the estimated cost for the phase (nullable string, e.g. "18818.0").
	EstimatedCost *string `json:"estimated_cost,omitempty"`

	// EstimatedHours is the estimated hours for the phase (nullable string).
	EstimatedHours *string `json:"estimated_hours,omitempty"`

	// IsBillable indicates whether time logged to this phase is billable.
	IsBillable bool `json:"is_billable,omitempty"`

	// IsBudget indicates whether this is a budget phase.
	IsBudget bool `json:"is_budget,omitempty"`

	// IsDefault indicates whether this is the default phase.
	IsDefault bool `json:"is_default,omitempty"`

	// IsMain indicates whether this is the main phase.
	IsMain bool `json:"is_main,omitempty"`

	// IsArchived indicates whether the phase is archived.
	IsArchived bool `json:"is_archived,omitempty"`

	// ParentID is the ID of the parent phase (nullable, for sub-phases).
	ParentID *int `json:"parent_id,omitempty"`

	// StartDate is the planned start date (nullable string, e.g. "2025-10-30").
	StartDate *string `json:"start_date,omitempty"`

	// EndDate is the planned end date (nullable string, e.g. "2025-10-30").
	EndDate *string `json:"end_date,omitempty"`

	// RateGroupID is the rate group associated with the phase.
	RateGroupID int `json:"rate_group_id,omitempty"`

	// RateMultiplier is the billing rate multiplier (string, e.g. "1.0").
	RateMultiplier string `json:"rate_multiplier,omitempty"`

	// CostRateMultiplier is the cost rate multiplier (string, e.g. "1.0").
	CostRateMultiplier string `json:"cost_rate_multiplier,omitempty"`

	// ProfitPercentage is the profit percentage for the phase (string, e.g. "0.0").
	ProfitPercentage string `json:"profit_percentage,omitempty"`

	// ProfitCenter is the profit center for the phase.
	ProfitCenter string `json:"profit_center,omitempty"`

	// BillRateType is the bill rate type for the phase.
	BillRateType string `json:"bill_rate_type,omitempty"`

	// ChildrenIDs is a list of child phase IDs.
	ChildrenIDs []int `json:"children_ids,omitempty"`

	// BudgetPhaseBy describes how the phase is budgeted.
	BudgetPhaseBy string `json:"budget_phase_by,omitempty"`

	// BudgetFixedFeeWith describes fixed fee budget configuration.
	BudgetFixedFeeWith string `json:"budget_fixed_fee_with,omitempty"`

	// BudgetHourlyWith describes hourly budget configuration.
	BudgetHourlyWith string `json:"budget_hourly_with,omitempty"`

	// BudgetInternalWith describes internal budget configuration.
	BudgetInternalWith string `json:"budget_internal_with,omitempty"`

	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// phaseListResponse is the wrapper for phase list API responses.
// The Mosaic API wraps phase arrays in {"phase": [...]}.
type phaseListResponse struct {
	Phase []Phase `json:"phase"`
}

// PhaseFilter contains filter parameters for listing phases by project.
type PhaseFilter struct {
	// ProjectID filters phases by project. Uses GET /api/{team_id}/phase?project_id=X.
	ProjectID int
}

// toQuery converts PhaseFilter into URL query parameter values.
func (f PhaseFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.ProjectID > 0 {
		q.Set("project_id", strconv.Itoa(f.ProjectID))
	}
	return q
}

// PhaseIndexFilter contains filter parameters for listing all phases via the index endpoint.
type PhaseIndexFilter struct {
	ListParams

	// IsBudget filters by budget phase status.
	IsBudget *bool

	// SearchText filters phases by name search.
	SearchText string

	// ProjectIDs filters by specific project IDs.
	ProjectIDs []int

	// PhaseIDs filters by specific phase IDs.
	PhaseIDs []int

	// ArchivedAfterDate filters phases archived after this date.
	ArchivedAfterDate *Date

	// ArchivedBeforeDate filters phases archived before this date.
	ArchivedBeforeDate *Date
}

// toQuery converts PhaseIndexFilter into URL query parameter values.
func (f PhaseIndexFilter) toQuery() url.Values {
	q := f.ListParams.toQuery()
	if f.IsBudget != nil {
		q.Set("is_budget", strconv.FormatBool(*f.IsBudget))
	}
	if f.SearchText != "" {
		q.Set("search_text", f.SearchText)
	}
	for _, id := range f.ProjectIDs {
		q.Add("project_ids", strconv.Itoa(id))
	}
	for _, id := range f.PhaseIDs {
		q.Add("phase_ids", strconv.Itoa(id))
	}
	if f.ArchivedAfterDate != nil && !f.ArchivedAfterDate.IsZero() {
		q.Set("archived_after_date", f.ArchivedAfterDate.String())
	}
	if f.ArchivedBeforeDate != nil && !f.ArchivedBeforeDate.IsZero() {
		q.Set("archived_before_date", f.ArchivedBeforeDate.String())
	}
	return q
}

// ListPhases retrieves phases for a specific project.
// Uses GET /api/{team_id}/phase?project_id=X.
func (c *Client) ListPhases(ctx context.Context, filter PhaseFilter) ([]Phase, error) {
	var resp phaseListResponse
	path := c.apiPath("phase")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing phases: %w", err)
	}
	return resp.Phase, nil
}

// ListAllPhases retrieves all phases across projects using the index endpoint.
// Uses GET /api/{team_id}/phase/index.
func (c *Client) ListAllPhases(ctx context.Context, filter PhaseIndexFilter) ([]Phase, error) {
	var resp phaseListResponse
	path := c.apiPath("phase", "index")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing all phases: %w", err)
	}
	return resp.Phase, nil
}

// CreatePhase creates a new phase and returns the created phase.
// Uses POST /api/{team_id}/phase.
func (c *Client) CreatePhase(ctx context.Context, phase *Phase) (*Phase, error) {
	var created Phase
	path := c.apiPath("phase")
	if err := c.post(ctx, path, phase, &created); err != nil {
		return nil, fmt.Errorf("creating phase: %w", err)
	}
	return &created, nil
}

// UpdatePhase updates an existing phase and returns the updated phase.
// Uses PUT /api/{team_id}/phase/{phase_id}.
func (c *Client) UpdatePhase(ctx context.Context, phaseID int, phase *Phase) (*Phase, error) {
	var updated Phase
	path := c.apiPath("phase", strconv.Itoa(phaseID))
	if err := c.put(ctx, path, phase, &updated); err != nil {
		return nil, fmt.Errorf("updating phase %d: %w", phaseID, err)
	}
	return &updated, nil
}

// phaseDeleteRequest is the request body for deleting a phase.
type phaseDeleteRequest struct {
	ProjectID    int  `json:"project_id"`
	ForceArchive bool `json:"force_archive,omitempty"`
	ForceDestroy bool `json:"force_destroy,omitempty"`
	Recursive    bool `json:"recursive,omitempty"`
}

// DeletePhase deletes a phase by ID.
// Uses DELETE /api/{team_id}/phase/{id}.
func (c *Client) DeletePhase(ctx context.Context, id, projectID int, forceArchive, forceDestroy, recursive bool) error {
	path := c.apiPath("phase", strconv.Itoa(id))
	body := phaseDeleteRequest{
		ProjectID:    projectID,
		ForceArchive: forceArchive,
		ForceDestroy: forceDestroy,
		Recursive:    recursive,
	}
	if err := c.deleteWithBody(ctx, path, body); err != nil {
		return fmt.Errorf("deleting phase %d: %w", id, err)
	}
	return nil
}

// ListPhasesByProject retrieves all phases belonging to a specific project.
func (c *Client) ListPhasesByProject(ctx context.Context, projectID int) ([]Phase, error) {
	return c.ListPhases(ctx, PhaseFilter{ProjectID: projectID})
}
