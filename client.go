package mosaic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// defaultBaseURL is the default Mosaic API base URL.
	defaultBaseURL = "https://api-server.prod.prod.us-east-1.mosaicapp.com"

	// defaultTenant is the default x-tenant header value.
	defaultTenant = "prod"

	// defaultRealmID is the default x-realm-id header value.
	defaultRealmID = "prod"
)

// Client is an API client for the Mosaic time management platform.
type Client struct {
	baseURL    string
	apiKey     string
	teamID     string
	origin     string
	tenant     string
	realmID    string
	httpClient *http.Client
}

// NewClient creates a new Mosaic API client with the given API key and team ID.
// Additional configuration can be provided via functional options.
func NewClient(apiKey, teamID string, opts ...Option) *Client {
	c := &Client{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		teamID:  teamID,
		tenant:  defaultTenant,
		realmID: defaultRealmID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// apiPath constructs the full API path for a given data type and optional sub-path segments.
func (c *Client) apiPath(dataType string, segments ...string) string {
	path := fmt.Sprintf("/api/%s/%s", c.teamID, dataType)
	for _, seg := range segments {
		path += "/" + seg
	}
	return path
}

// maxRetries is the number of times to retry a request after a rate limit.
const maxRetries = 3

// minRetryDelay is the minimum delay between retries when the server returns
// a Retry-After of 0 seconds.
const minRetryDelay = 2 * time.Second

// do executes an HTTP request against the Mosaic API, handling authentication
// headers, JSON encoding/decoding, error mapping, and automatic retry on rate
// limit (429) responses.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, result any) error {
	// Marshal the body once so we can replay it on retries.
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
	}

	for attempt := range maxRetries + 1 {
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.baseURL, "/")+path, bodyReader)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		// Set authentication and required multi-tenant headers.
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("x-tenant", c.tenant)
		req.Header.Set("x-realm-id", c.realmID)
		if c.origin != "" {
			req.Header.Set("Origin", c.origin)
		}

		// Set query parameters.
		if len(query) > 0 {
			q := req.URL.Query()
			for k, vals := range query {
				for _, v := range vals {
					q.Add(k, v)
				}
			}
			req.URL.RawQuery = q.Encode()
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("executing request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}

		if resp.StatusCode >= 400 {
			apiErr := c.handleErrorResponse(resp, respBody)

			// Auto-retry on rate limit if we have attempts left.
			var rlErr *RateLimitError
			if errors.As(apiErr, &rlErr) && attempt < maxRetries {
				delay := rlErr.RetryAfter
				if delay < minRetryDelay {
					delay = minRetryDelay
				}
				slog.Warn("rate limited, retrying",
					"attempt", attempt+1,
					"max_retries", maxRetries,
					"retry_after", delay,
					"path", path,
				)
				select {
				case <-time.After(delay):
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return apiErr
		}

		// For DELETE with 204 No Content, skip decoding.
		if resp.StatusCode == http.StatusNoContent || result == nil {
			return nil
		}

		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}

		return nil
	}

	// Should not be reached, but just in case.
	return fmt.Errorf("exhausted retries for %s %s", method, path)
}

// handleErrorResponse maps an HTTP error response to the appropriate error type.
func (c *Client) handleErrorResponse(resp *http.Response, body []byte) error {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Body:       body,
	}

	// Try to extract a message from the response body.
	var errBody struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(body, &errBody); err == nil {
		if errBody.Message != "" {
			apiErr.Message = errBody.Message
		} else if errBody.Error != "" {
			apiErr.Message = errBody.Error
		}
	}

	// Handle rate limiting specially.
	if resp.StatusCode == http.StatusTooManyRequests {
		rateLimitErr := &RateLimitError{Err: apiErr}
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				rateLimitErr.RetryAfter = time.Duration(seconds) * time.Second
			}
		}
		return rateLimitErr
	}

	return apiErr
}

// get performs a GET request against the Mosaic API.
func (c *Client) get(ctx context.Context, path string, query url.Values, result any) error {
	return c.do(ctx, http.MethodGet, path, query, nil, result)
}

// post performs a POST request against the Mosaic API.
func (c *Client) post(ctx context.Context, path string, body any, result any) error {
	return c.do(ctx, http.MethodPost, path, nil, body, result)
}

// put performs a PUT request against the Mosaic API.
func (c *Client) put(ctx context.Context, path string, body any, result any) error {
	return c.do(ctx, http.MethodPut, path, nil, body, result)
}

// patchRequest performs a PATCH request against the Mosaic API.
func (c *Client) patchRequest(ctx context.Context, path string, body any, result any) error {
	return c.do(ctx, http.MethodPatch, path, nil, body, result)
}

// deleteNoBody performs a DELETE request against the Mosaic API without a body.
func (c *Client) deleteNoBody(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil, nil)
}

// deleteWithBody performs a DELETE request against the Mosaic API with a request body.
func (c *Client) deleteWithBody(ctx context.Context, path string, body any) error {
	return c.do(ctx, http.MethodDelete, path, nil, body, nil)
}

// deleteWithQuery performs a DELETE request against the Mosaic API with query parameters.
func (c *Client) deleteWithQuery(ctx context.Context, path string, query url.Values) error {
	return c.do(ctx, http.MethodDelete, path, query, nil, nil)
}

// RawGet performs a GET request against an arbitrary API path and decodes the
// response into result. The path should be the full path starting with /api/
// (the base URL is prepended automatically). This is useful for debugging and
// exploring API responses.
func (c *Client) RawGet(ctx context.Context, path string, query url.Values, result any) error {
	return c.do(ctx, http.MethodGet, path, query, nil, result)
}
