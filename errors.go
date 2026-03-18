package mosaic

import (
	"errors"
	"fmt"
	"time"
)

// Sentinel errors for common API error conditions.
var (
	// ErrNotFound is returned when the requested resource does not exist.
	ErrNotFound = errors.New("mosaic: resource not found")

	// ErrUnauthorized is returned when the API key is invalid or missing.
	ErrUnauthorized = errors.New("mosaic: unauthorized")

	// ErrForbidden is returned when the caller lacks permission for the operation.
	ErrForbidden = errors.New("mosaic: forbidden")

	// ErrRateLimited is returned when the API rate limit has been exceeded.
	ErrRateLimited = errors.New("mosaic: rate limit exceeded")

	// ErrBadRequest is returned when the request is malformed.
	ErrBadRequest = errors.New("mosaic: bad request")
)

// APIError represents an error response from the Mosaic API.
type APIError struct {
	// StatusCode is the HTTP status code from the response.
	StatusCode int

	// Message is a human-readable error description.
	Message string

	// Body is the raw response body.
	Body []byte
}

// Error returns a string representation of the API error.
func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("mosaic: api error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("mosaic: api error (status %d): %s", e.StatusCode, string(e.Body))
}

// Unwrap returns the corresponding sentinel error for the status code, if any.
func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case 400:
		return ErrBadRequest
	case 401:
		return ErrUnauthorized
	case 403:
		return ErrForbidden
	case 404:
		return ErrNotFound
	case 429:
		return ErrRateLimited
	default:
		return nil
	}
}

// RateLimitError is returned when the API rate limit has been exceeded.
// It includes the duration after which the request can be retried.
type RateLimitError struct {
	// RetryAfter is the duration to wait before retrying the request.
	RetryAfter time.Duration

	// Err is the underlying API error.
	Err *APIError
}

// Error returns a string representation of the rate limit error.
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("mosaic: rate limit exceeded, retry after %s", e.RetryAfter)
}

// Unwrap returns the sentinel ErrRateLimited error.
func (e *RateLimitError) Unwrap() error {
	return ErrRateLimited
}
