package mosaic

import (
	"context"
	"fmt"
	"strconv"
)

// Office represents a physical office location in Mosaic.
type Office struct {
	// MosaicID is the unique identifier for the office in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this office belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// Name is the display name of the office.
	Name string `json:"name"`

	// Address1 is the first line of the office address.
	Address1 string `json:"address_1,omitempty"`

	// Address2 is the second line of the office address.
	Address2 string `json:"address_2,omitempty"`

	// Address3 is the third line of the office address.
	Address3 string `json:"address_3,omitempty"`

	// City is the city where the office is located.
	City string `json:"city,omitempty"`

	// State is the state or province where the office is located.
	State string `json:"state,omitempty"`

	// Zip is the postal/ZIP code for the office.
	Zip string `json:"zip,omitempty"`

	// Country is the country where the office is located.
	Country string `json:"country,omitempty"`

	// CreatedAt is the timestamp when the office was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the office was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// officeListResponse is the wrapper for the GET /api/{team_id}/offices response.
type officeListResponse struct {
	Offices []Office `json:"offices"`
}

// ListOffices retrieves all offices.
// Uses GET /api/{team_id}/offices.
func (c *Client) ListOffices(ctx context.Context) ([]Office, error) {
	var resp officeListResponse
	path := c.apiPath("offices")
	if err := c.get(ctx, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing offices: %w", err)
	}
	return resp.Offices, nil
}

// CreateOffice creates a new office and returns the created office.
// Uses POST /api/{team_id}/offices.
func (c *Client) CreateOffice(ctx context.Context, office *Office) (*Office, error) {
	var created Office
	path := c.apiPath("offices")
	if err := c.post(ctx, path, office, &created); err != nil {
		return nil, fmt.Errorf("creating office: %w", err)
	}
	return &created, nil
}

// UpdateOffice updates an existing office and returns the updated office.
// Uses PUT /api/{team_id}/offices/{id}.
func (c *Client) UpdateOffice(ctx context.Context, id int, office *Office) (*Office, error) {
	var updated Office
	path := c.apiPath("offices", strconv.Itoa(id))
	if err := c.put(ctx, path, office, &updated); err != nil {
		return nil, fmt.Errorf("updating office %d: %w", id, err)
	}
	return &updated, nil
}

// DeleteOffice deletes an office by ID.
// Uses DELETE /api/{team_id}/offices/{id}.
func (c *Client) DeleteOffice(ctx context.Context, id int) error {
	path := c.apiPath("offices", strconv.Itoa(id))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting office %d: %w", id, err)
	}
	return nil
}
