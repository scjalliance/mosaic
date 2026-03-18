package mosaic

import (
	"context"
	"fmt"
	"strconv"
)

// ClientEntity represents a client organization in Mosaic.
// Named ClientEntity to avoid collision with the API Client type.
type ClientEntity struct {
	// MosaicID is the unique identifier for the client in Mosaic.
	MosaicID int `json:"mosaic_id,omitempty"`

	// MosaicTeamID is the team ID this client belongs to.
	MosaicTeamID int `json:"mosaic_team_id,omitempty"`

	// ClientName is the client's display name.
	ClientName string `json:"client_name"`

	// Description is the client's description.
	Description string `json:"description,omitempty"`

	// IsArchived indicates whether the client has been archived.
	IsArchived bool `json:"is_archived,omitempty"`

	// CreatedAt is the timestamp when the client was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the timestamp when the client was last updated.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// clientListResponse is the wrapper for the GET /api/{team_id}/client response,
// which returns {"client": [...]}.
type clientListResponse struct {
	Client []ClientEntity `json:"client"`
}

// ListClients retrieves all clients.
// Uses GET /api/{team_id}/client.
func (c *Client) ListClients(ctx context.Context) ([]ClientEntity, error) {
	var resp clientListResponse
	path := c.apiPath("client")
	if err := c.get(ctx, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("listing clients: %w", err)
	}
	return resp.Client, nil
}

// CreateClient creates a new client and returns the created client.
// Uses POST /api/{team_id}/client.
func (c *Client) CreateClient(ctx context.Context, client *ClientEntity) (*ClientEntity, error) {
	var created ClientEntity
	path := c.apiPath("client")
	if err := c.post(ctx, path, client, &created); err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}
	return &created, nil
}

// UpdateClient updates an existing client and returns the updated client.
// Uses PUT /api/{team_id}/client/{client_id}. Note: there is no DELETE endpoint for clients.
func (c *Client) UpdateClient(ctx context.Context, clientID int, client *ClientEntity) (*ClientEntity, error) {
	var updated ClientEntity
	path := c.apiPath("client", strconv.Itoa(clientID))
	if err := c.put(ctx, path, client, &updated); err != nil {
		return nil, fmt.Errorf("updating client %d: %w", clientID, err)
	}
	return &updated, nil
}
