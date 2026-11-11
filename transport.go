package mosaic

import (
	"net"
	"net/http"
	"time"
)

// newDefaultHTTPClient returns the http.Client NewClient uses, built from
// base (normally http.DefaultTransport), with defaultTimeout for a whole
// request.
//
//   - An *http.Transport is cloned, keeping the program's proxy, dial, and TLS
//     settings without changing the shared global. If the program set no
//     ResponseHeaderTimeout, the clone gets defaultHeaderTimeout; a value the
//     program chose is kept. That timer starts once the request is written,
//     so a connect or TLS failure keeps its own error.
//   - Any other RoundTripper (an egress policy, proxy, or instrumentation a
//     program installed) is used as is, so Mosaic requests and their bearer
//     token still go through it. It gets only defaultTimeout; pass
//     WithHTTPClient to set a header timeout on it.
//   - A nil base, or a nil *http.Transport, gets a new transport with the
//     stock limits and the header timeout.
func newDefaultHTTPClient(base http.RoundTripper) *http.Client {
	t, isTransport := base.(*http.Transport)
	switch {
	case base == nil || (isTransport && t == nil):
		t = stockTransport()
	case isTransport:
		t = t.Clone()
	default:
		return &http.Client{Transport: base, Timeout: defaultTimeout}
	}
	if t.ResponseHeaderTimeout == 0 {
		t.ResponseHeaderTimeout = defaultHeaderTimeout
	}
	return &http.Client{Transport: t, Timeout: defaultTimeout}
}

// stockTransport returns a transport with the limits http.DefaultTransport
// ships with.
func stockTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
