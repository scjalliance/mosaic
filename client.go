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
	pacer      *pacer // never nil; see WithRequestInterval
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
		pacer: &pacer{interval: DefaultRequestInterval},
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

// minRetryDelay is the least time a 429 holds back requests, even when
// Retry-After is shorter. A 429 with no usable Retry-After waits
// rateLimitPause. See parseRetryAfter.
const minRetryDelay = 2 * time.Second

// maxRetryAfter caps the pause a single 429 can ask for. Mosaic's window is
// 5 minutes, so a longer wait is never needed.
const maxRetryAfter = 5 * time.Minute

// do executes an HTTP request against the Mosaic API, handling authentication
// headers, JSON encoding/decoding, error mapping, and automatic retry on rate
// limit (429) responses.
//
// Every attempt first waits its turn on the Client's pacer, so requests start
// at least the request interval apart. A 429 pauses the pacer, which holds
// back every request from the Client, not only the one that got the 429.
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

	var lastRateLimit *RateLimitError // kept so a timeout during its pause says why
	for attempt := range maxRetries + 1 {
		if err := c.pacer.wait(ctx); err != nil {
			if lastRateLimit != nil {
				return fmt.Errorf("%w while waiting out a rate limit: %w", err, lastRateLimit)
			}
			return err
		}

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

			// On a 429, hold back every request from this Client, then retry
			// this one if it has attempts left. The pause applies to the last
			// attempt too, so the next caller does not walk into the limit.
			var rlErr *RateLimitError
			if errors.As(apiErr, &rlErr) {
				lastRateLimit = rlErr
				delay := rlErr.RetryAfter // already within its bounds
				c.pacer.pause(delay)
				if attempt < maxRetries {
					slog.Warn("rate limited, retrying",
						"attempt", attempt+1,
						"max_retries", maxRetries,
						"retry_after", delay,
						"path", path,
					)
					continue // the pacer wait at the top sleeps out the pause
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
		// Measure an HTTP-date Retry-After against the server's own clock
		// (its Date header) when it sent one, so local clock skew does not
		// shorten the pause.
		now := time.Now()
		if d, err := http.ParseTime(resp.Header.Get("Date")); err == nil {
			now = d
		}
		rateLimitErr := &RateLimitError{Err: apiErr, RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"), now)}
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

// RawPut performs a PUT request against an arbitrary API path, sending body as
// the JSON request payload and decoding the response into result. The path
// should be the full path starting with /api/ (the base URL is prepended
// automatically). This is useful for round-tripping a resource's raw JSON
// representation (read-modify-write) without modeling every field in a typed
// struct, so fields not covered by a struct are preserved. Pass a nil result to
// ignore the response body.
func (c *Client) RawPut(ctx context.Context, path string, body any, result any) error {
	return c.do(ctx, http.MethodPut, path, nil, body, result)
}

// parseRetryAfter reads a Retry-After header value, either delay-seconds or
// an HTTP date measured from now, as the pause a 429 asks for. A missing or
// unreadable value gives rateLimitPause. The result is kept between
// minRetryDelay and maxRetryAfter, so a 429 always holds requests back for a
// real interval and one odd header cannot stall the Client for hours.
func parseRetryAfter(v string, now time.Time) time.Duration {
	return min(max(rawRetryAfter(v, now), minRetryDelay), maxRetryAfter)
}

// rawRetryAfter is parseRetryAfter before the bounds are applied, except that
// it caps values too large to hold in a time.Duration.
func rawRetryAfter(v string, now time.Time) time.Duration {
	v = strings.TrimSpace(v)
	seconds, err := strconv.ParseInt(v, 10, 64)
	if errors.Is(err, strconv.ErrRange) && seconds > 0 {
		return maxRetryAfter // too many digits for int64, still a long wait
	}
	if err == nil && seconds >= 0 {
		return time.Duration(min(seconds, int64(maxRetryAfter/time.Second))) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		return min(max(t.Sub(now), 0), maxRetryAfter)
	}
	return rateLimitPause
}
