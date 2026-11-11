package mosaic

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

// newDefaultHTTPClient returns the http.Client NewClient uses: base wrapped
// so a response's headers must arrive within defaultHeaderTimeout, and
// defaultTimeout for the whole request.
//
// base is normally http.DefaultTransport. Whatever it is (the stock
// transport, or an egress policy, proxy, or instrumentation RoundTripper a
// program installed), Mosaic requests and their bearer token go through it.
// A nil base gets a transport with the stock limits.
func newDefaultHTTPClient(base http.RoundTripper) *http.Client {
	if t, ok := base.(*http.Transport); base == nil || (ok && t == nil) {
		base = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}
	return &http.Client{
		Transport: &headerTimeoutTransport{base: base, timeout: defaultHeaderTimeout},
		Timeout:   defaultTimeout,
	}
}

// headerTimeoutTransport fails a request whose response headers do not
// arrive within timeout, like http.Transport.ResponseHeaderTimeout, but
// around any RoundTripper. Once headers arrive, reading the body is not
// limited here.
type headerTimeoutTransport struct {
	base    http.RoundTripper
	timeout time.Duration
}

// RoundTrip sends req through t.base, cancelling it with
// errHeaderTimeout if headers take longer than t.timeout.
func (t *headerTimeoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancelCause(req.Context())
	timer := time.AfterFunc(t.timeout, func() { cancel(errHeaderTimeout) })
	resp, err := t.base.RoundTrip(req.WithContext(ctx))
	if !timer.Stop() {
		// The timer fired: the request was cancelled for taking too long,
		// whatever error the base reported.
		if resp != nil {
			resp.Body.Close()
		}
		cancel(nil)
		return nil, errHeaderTimeout
	}
	if err != nil {
		cancel(nil)
		return nil, err
	}
	// Keep the context alive while the caller reads the body; release it on
	// Close.
	resp.Body = &cancelOnClose{ReadCloser: resp.Body, cancel: func() { cancel(nil) }}
	return resp, nil
}

// cancelOnClose cancels a request's context when its body is closed.
type cancelOnClose struct {
	io.ReadCloser
	cancel func()
}

// Close closes the body and then releases the request's context.
func (c *cancelOnClose) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

// errHeaderTimeout is returned when response headers do not arrive in time.
// It is a net.Error that reports a timeout, so isTimeout treats it like
// Mosaic's own 504.
var errHeaderTimeout net.Error = headerTimeoutError{}

// headerTimeoutError is the type of errHeaderTimeout.
type headerTimeoutError struct{}

// Error describes the timeout.
func (headerTimeoutError) Error() string { return "mosaic: timeout awaiting response headers" }

// Timeout reports true: this is a timeout.
func (headerTimeoutError) Timeout() bool { return true }

// Temporary reports true; it is part of net.Error.
func (headerTimeoutError) Temporary() bool { return true }
