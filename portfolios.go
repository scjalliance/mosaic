package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Portfolio represents a portfolio (grouping of projects) in Mosaic.
// Field names correspond to the snake_case JSON fields used by the Mosaic API.
// The API calls these "boards" internally.
type Portfolio struct {
	// MosaicID is the unique identifier for the portfolio.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team this portfolio belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Name is the display name of the portfolio.
	Name string `json:"name"`

	// Color is the hex color code for the portfolio.
	Color string `json:"color,omitempty"`

	// IsPrivate indicates whether the portfolio is private.
	IsPrivate bool `json:"is_private"`

	// IsAdministrative indicates whether this is an administrative portfolio.
	IsAdministrative bool `json:"is_administrative,omitempty"`

	// IsIntegration indicates whether this portfolio was created by an integration.
	IsIntegration bool `json:"is_integration,omitempty"`

	// IsPersonal indicates whether this is a personal portfolio.
	IsPersonal bool `json:"is_personal,omitempty"`

	// Archive indicates whether to archive the portfolio (used in PUT).
	Archive bool `json:"archive,omitempty"`

	// Unarchive indicates whether to unarchive the portfolio (used in PUT).
	Unarchive bool `json:"unarchive,omitempty"`

	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// portfolioIndexResponse is the wrapper object returned by the portfolio index endpoint.
type portfolioIndexResponse struct {
	Boards []Portfolio `json:"boards"`
}

// GetPortfolio retrieves a single portfolio by ID.
// Uses GET /api/{team_id}/portfolio?portfolio_id=X.
func (c *Client) GetPortfolio(ctx context.Context, portfolioID int) (*Portfolio, error) {
	var portfolio Portfolio
	path := c.apiPath("portfolio")
	q := make(url.Values)
	q.Set("portfolio_id", strconv.Itoa(portfolioID))
	if err := c.get(ctx, path, q, &portfolio); err != nil {
		return nil, fmt.Errorf("getting portfolio %d: %w", portfolioID, err)
	}
	return &portfolio, nil
}

// ListAllPortfolios retrieves all portfolios using the index endpoint.
// Uses GET /api/{team_id}/portfolio/index.
// The API returns portfolios wrapped in a {"boards": [...]} object.
func (c *Client) ListAllPortfolios(ctx context.Context) ([]Portfolio, error) {
	var resp portfolioIndexResponse
	path := c.apiPath("portfolio", "index")
	if err := c.get(ctx, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing all portfolios: %w", err)
	}
	return resp.Boards, nil
}

// CreatePortfolio creates a new portfolio and returns the created portfolio.
// Uses POST /api/{team_id}/portfolio.
func (c *Client) CreatePortfolio(ctx context.Context, portfolio *Portfolio) (*Portfolio, error) {
	var created Portfolio
	path := c.apiPath("portfolio")
	if err := c.post(ctx, path, portfolio, &created); err != nil {
		return nil, fmt.Errorf("creating portfolio: %w", err)
	}
	return &created, nil
}

// UpdatePortfolio updates an existing portfolio.
// Uses PUT /api/{team_id}/portfolio/{portfolio_id}.
func (c *Client) UpdatePortfolio(ctx context.Context, portfolioID int, portfolio *Portfolio) (*Portfolio, error) {
	var updated Portfolio
	path := c.apiPath("portfolio", strconv.Itoa(portfolioID))
	if err := c.put(ctx, path, portfolio, &updated); err != nil {
		return nil, fmt.Errorf("updating portfolio %d: %w", portfolioID, err)
	}
	return &updated, nil
}

// portfolioDeleteRequest is the request body for deleting a portfolio.
type portfolioDeleteRequest struct {
	HardDelete bool `json:"hard_delete,omitempty"`
}

// DeletePortfolio deletes a portfolio by ID.
// Uses DELETE /api/{team_id}/portfolio/{portfolio_id}.
func (c *Client) DeletePortfolio(ctx context.Context, portfolioID int, hardDelete bool) error {
	path := c.apiPath("portfolio", strconv.Itoa(portfolioID))
	body := portfolioDeleteRequest{HardDelete: hardDelete}
	if err := c.deleteWithBody(ctx, path, body); err != nil {
		return fmt.Errorf("deleting portfolio %d: %w", portfolioID, err)
	}
	return nil
}
