package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Project represents a project in Mosaic.
// Field names correspond to the snake_case JSON fields used by the Mosaic API.
type Project struct {
	// MosaicID is the unique identifier for the project.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID the project belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Title is the display name of the project.
	Title string `json:"title"`

	// Description is the project description (nullable).
	Description *string `json:"description,omitempty"`

	// ProjectNumber is the project number (e.g., "25-000715").
	ProjectNumber string `json:"project_number,omitempty"`

	// ClientID is the ID of the client associated with the project.
	ClientID int `json:"client_id,omitempty"`

	// PortfolioID is the ID of the portfolio this project belongs to.
	PortfolioID int `json:"portfolio_id,omitempty"`

	// BudgetStatus is the budget status of the project (e.g., "active", "archived").
	BudgetStatus string `json:"budget_status,omitempty"`

	// Total is the total budget amount for the project (nullable, e.g., "80120.0").
	Total *string `json:"total,omitempty"`

	// Fee is the fee amount for the project (nullable, e.g., "80120.0").
	Fee *string `json:"fee,omitempty"`

	// CurrencyCode is the project's currency code.
	CurrencyCode string `json:"currency_code,omitempty"`

	// IsBillable indicates whether time entries are billable by default.
	IsBillable bool `json:"is_billable,omitempty"`

	// IsMain indicates whether this is the main project.
	IsMain bool `json:"is_main,omitempty"`

	// IsArchived indicates whether the project is archived.
	IsArchived bool `json:"is_archived,omitempty"`

	// StartDate is the planned start date of the project (nullable, e.g., "2025-10-30").
	StartDate *string `json:"start_date,omitempty"`

	// EndDate is the planned end date of the project (nullable, e.g., "2025-10-30").
	EndDate *string `json:"end_date,omitempty"`

	// ProfitCenter is the profit center for the project.
	ProfitCenter string `json:"profit_center,omitempty"`

	// BudgetBill is the budget bill amount (nullable).
	BudgetBill *string `json:"budget_bill,omitempty"`

	// BudgetCost is the budget cost amount (nullable).
	BudgetCost *string `json:"budget_cost,omitempty"`

	// BudgetHours is the budget hours (nullable).
	BudgetHours *string `json:"budget_hours,omitempty"`

	// EstimatedCost is the estimated cost of the project (nullable).
	EstimatedCost *string `json:"estimated_cost,omitempty"`

	// EstimatedHours is the estimated hours for the project (nullable).
	EstimatedHours *string `json:"estimated_hours,omitempty"`

	// RequireVaccination indicates whether vaccination is required.
	RequireVaccination bool `json:"require_vaccination,omitempty"`

	// MainPhaseID is the ID of the main phase (nullable).
	MainPhaseID *int `json:"main_phase_id,omitempty"`

	// CreatedAt is the timestamp when the project was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the project was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// projectListResponse is the wrapper for the project list API response.
// The API returns {"project": [...], "total_count": N}.
type projectListResponse struct {
	Project    []Project `json:"project"`
	TotalCount int       `json:"total_count"`
}

// ProjectFilter contains filter parameters for listing projects.
type ProjectFilter struct {
	ListParams

	// PortfolioID filters projects by portfolio.
	PortfolioID int

	// BudgetStatus filters projects by budget status.
	BudgetStatus string

	// IsArchived filters projects by archived status.
	IsArchived *bool

	// IsMain filters projects by main status.
	IsMain *bool

	// ProjectIDs filters to specific project IDs.
	ProjectIDs []int

	// ArchivedAfterDate filters projects archived after this date.
	ArchivedAfterDate *Date

	// ArchivedBeforeDate filters projects archived before this date.
	ArchivedBeforeDate *Date
}

// toQuery converts ProjectFilter into URL query parameter values.
func (f ProjectFilter) toQuery() url.Values {
	q := f.ListParams.toQuery()
	if f.PortfolioID > 0 {
		q.Set("portfolio_id", strconv.Itoa(f.PortfolioID))
	}
	if f.BudgetStatus != "" {
		q.Set("budget_status", f.BudgetStatus)
	}
	if f.IsArchived != nil {
		q.Set("is_archived", strconv.FormatBool(*f.IsArchived))
	}
	if f.IsMain != nil {
		q.Set("is_main", strconv.FormatBool(*f.IsMain))
	}
	for _, id := range f.ProjectIDs {
		q.Add("project_ids[]", strconv.Itoa(id))
	}
	if f.ArchivedAfterDate != nil && !f.ArchivedAfterDate.IsZero() {
		q.Set("archived_after_date", f.ArchivedAfterDate.String())
	}
	if f.ArchivedBeforeDate != nil && !f.ArchivedBeforeDate.IsZero() {
		q.Set("archived_before_date", f.ArchivedBeforeDate.String())
	}
	return q
}

// ListProjects retrieves a list of projects matching the given filter.
// Uses GET /api/{team_id}/project.
// NOTE: The API returns 404 if no portfolio_id filter is provided.
func (c *Client) ListProjects(ctx context.Context, filter ProjectFilter) ([]Project, error) {
	var resp projectListResponse
	path := c.apiPath("project")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	return resp.Project, nil
}

// CreateProject creates a new project and returns the created project.
// Uses POST /api/{team_id}/project.
func (c *Client) CreateProject(ctx context.Context, project *Project) (*Project, error) {
	var created Project
	path := c.apiPath("project")
	if err := c.post(ctx, path, project, &created); err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}
	return &created, nil
}

// UpdateProject updates an existing project and returns the updated project.
// Uses PUT /api/{team_id}/project/{project_id}.
func (c *Client) UpdateProject(ctx context.Context, projectID int, project *Project) (*Project, error) {
	var updated Project
	path := c.apiPath("project", strconv.Itoa(projectID))
	if err := c.put(ctx, path, project, &updated); err != nil {
		return nil, fmt.Errorf("updating project %d: %w", projectID, err)
	}
	return &updated, nil
}

// projectDeleteRequest is the request body for deleting a project.
type projectDeleteRequest struct {
	ForceArchive bool `json:"force_archive,omitempty"`
	ForceDestroy bool `json:"force_destroy,omitempty"`
}

// DeleteProject deletes a project by ID.
// Uses DELETE /api/{team_id}/project/{id}.
func (c *Client) DeleteProject(ctx context.Context, id int, forceArchive, forceDestroy bool) error {
	path := c.apiPath("project", strconv.Itoa(id))
	body := projectDeleteRequest{ForceArchive: forceArchive, ForceDestroy: forceDestroy}
	if err := c.deleteWithBody(ctx, path, body); err != nil {
		return fmt.Errorf("deleting project %d: %w", id, err)
	}
	return nil
}

// TotalOverride contains the fields for overriding project fee/revenue/hours/cost.
type TotalOverride struct {
	// Fee overrides the project fee.
	Fee float64 `json:"fee,omitempty"`

	// BudgetBill overrides the budget bill amount.
	BudgetBill float64 `json:"budget_bill,omitempty"`

	// BudgetCost overrides the budget cost amount.
	BudgetCost float64 `json:"budget_cost,omitempty"`

	// BudgetHours overrides the budget hours.
	BudgetHours float64 `json:"budget_hours,omitempty"`
}

// OverrideProjectTotal overrides fee/revenue/hours/cost on a project.
// Uses PATCH /api/{team_id}/project/{project_id}/total_override.
func (c *Client) OverrideProjectTotal(ctx context.Context, projectID int, override TotalOverride) error {
	path := c.apiPath("project", strconv.Itoa(projectID), "total_override")
	if err := c.patchRequest(ctx, path, override, nil); err != nil {
		return fmt.Errorf("overriding project %d total: %w", projectID, err)
	}
	return nil
}

// ListProjectsByPortfolio retrieves all projects belonging to a specific portfolio.
func (c *Client) ListProjectsByPortfolio(ctx context.Context, portfolioID int, params ListParams) ([]Project, error) {
	filter := ProjectFilter{
		ListParams:  params,
		PortfolioID: portfolioID,
	}
	return c.ListProjects(ctx, filter)
}
