package mosaic

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithBaseURL sets a custom base URL for the Mosaic API.
// The default is https://api-server.prod.prod.us-east-1.mosaicapp.com.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithOrigin sets the Origin header value sent with each request.
func WithOrigin(origin string) Option {
	return func(c *Client) {
		c.origin = origin
	}
}

// WithTenant sets the x-tenant header value. The default is "prod".
func WithTenant(tenant string) Option {
	return func(c *Client) {
		c.tenant = tenant
	}
}

// WithRealmID sets the x-realm-id header value. The default is "prod".
func WithRealmID(realmID string) Option {
	return func(c *Client) {
		c.realmID = realmID
	}
}

// WithHTTPClient sets a custom *http.Client for making API requests.
// This is useful for configuring timeouts, transport settings, or testing. The
// default client times out a request after 75 seconds, a little longer than
// Mosaic's own 60-second limit.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithRequestInterval sets the least time between the starts of two requests
// from this Client. The default is [DefaultRequestInterval], which keeps a
// Client under Mosaic's rate limit. Zero or less turns spacing off; a 429
// still holds back all requests.
func WithRequestInterval(d time.Duration) Option {
	return func(c *Client) {
		c.pacer = &pacer{interval: max(d, 0)}
	}
}
