package mosaic

import (
	"context"
	"fmt"
	"net/url"
)

// RequestLog represents the status of an asynchronous API request in Mosaic.
// Some API operations are processed asynchronously; this resource allows
// checking on their status.
type RequestLog struct {
	// APIRequestID is the unique identifier for the API request.
	APIRequestID string `json:"api_request_id,omitempty"`

	// DataType is the type of data the request operates on.
	DataType string `json:"data_type,omitempty"`

	// Status is the current status of the request.
	Status string `json:"status,omitempty"`
}

// GetRequestLog retrieves the status of an asynchronous API request.
// Uses GET /api/{team_id}/request_logs.
// At least one of apiRequestID or dataType should be provided.
func (c *Client) GetRequestLog(ctx context.Context, apiRequestID, dataType string) (*RequestLog, error) {
	var resp RequestLog
	path := c.apiPath("request_logs")
	q := make(url.Values)
	if apiRequestID != "" {
		q.Set("api_request_id", apiRequestID)
	}
	if dataType != "" {
		q.Set("data_type", dataType)
	}
	if err := c.get(ctx, path, q, &resp); err != nil {
		return nil, fmt.Errorf("getting request log: %w", err)
	}
	return &resp, nil
}
