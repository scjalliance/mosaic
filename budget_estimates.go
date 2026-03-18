package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// BudgetEstimate represents a budget estimate in Mosaic.
// Field names correspond to the snake_case JSON fields used by the Mosaic API.
type BudgetEstimate struct {
	// ID is the unique identifier for the budget estimate.
	ID int `json:"id,omitempty"`

	// ProjectID is the ID of the associated project.
	ProjectID int `json:"project_id"`

	// PhaseID is the ID of the associated phase.
	PhaseID int `json:"phase_id,omitempty"`

	// MemberID is the ID of the associated member.
	MemberID int `json:"member_id,omitempty"`

	// RoleID is the ID of the associated role.
	RoleID int `json:"role_id,omitempty"`

	// ScopeID is the ID of the associated scope.
	ScopeID int `json:"scope_id,omitempty"`

	// StandardWorkCategoryID is the work category ID.
	StandardWorkCategoryID int `json:"standard_work_category_id,omitempty"`

	// EstimatedHours is the estimated number of hours.
	EstimatedHours float64 `json:"estimated_hours"`

	// EstimatedAmount is the estimated monetary amount.
	EstimatedAmount float64 `json:"estimated_amount"`

	// EstimatedPercentage is the estimated percentage.
	EstimatedPercentage float64 `json:"estimated_percentage"`

	// RateType is the rate type for the estimate.
	RateType string `json:"rate_type,omitempty"`
}

// BudgetEstimateFilter contains filter parameters for listing budget estimates.
type BudgetEstimateFilter struct {
	// ProjectID filters budget estimates by project.
	ProjectID int
}

// toQuery converts BudgetEstimateFilter into URL query parameter values.
func (f BudgetEstimateFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.ProjectID > 0 {
		q.Set("project_id", strconv.Itoa(f.ProjectID))
	}
	return q
}

// ListBudgetEstimates retrieves a list of budget estimates.
// Uses GET /api/{team_id}/budget_estimate/index.
func (c *Client) ListBudgetEstimates(ctx context.Context, filter BudgetEstimateFilter) ([]BudgetEstimate, error) {
	var resp []BudgetEstimate
	path := c.apiPath("budget_estimate", "index")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing budget estimates: %w", err)
	}
	return resp, nil
}

// CreateBudgetEstimate creates a new budget estimate and returns the created estimate.
// Uses POST /api/{team_id}/budget_estimate.
func (c *Client) CreateBudgetEstimate(ctx context.Context, estimate *BudgetEstimate) (*BudgetEstimate, error) {
	var created BudgetEstimate
	path := c.apiPath("budget_estimate")
	if err := c.post(ctx, path, estimate, &created); err != nil {
		return nil, fmt.Errorf("creating budget estimate: %w", err)
	}
	return &created, nil
}

// UpdateBudgetEstimate updates an existing budget estimate.
// Uses PUT /api/{team_id}/budget_estimate/{id}.
func (c *Client) UpdateBudgetEstimate(ctx context.Context, id int, estimate *BudgetEstimate) (*BudgetEstimate, error) {
	var updated BudgetEstimate
	path := c.apiPath("budget_estimate", strconv.Itoa(id))
	if err := c.put(ctx, path, estimate, &updated); err != nil {
		return nil, fmt.Errorf("updating budget estimate %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteBudgetEstimate deletes a budget estimate by ID.
// Uses DELETE /api/{team_id}/budget_estimate/{id}.
func (c *Client) DeleteBudgetEstimate(ctx context.Context, id int) error {
	path := c.apiPath("budget_estimate", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting budget estimate %d: %w", id, err)
	}
	return nil
}
