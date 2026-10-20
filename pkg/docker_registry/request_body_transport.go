package docker_registry

import (
	"fmt"
	"net/http"
)

var _ http.RoundTripper = (*requestBodyTransport)(nil)

type requestBodyTransport struct {
	underlying http.RoundTripper
}

func (t *requestBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body == nil || req.Body == http.NoBody || req.GetBody == nil {
		return t.underlying.RoundTrip(req)
	}

	// Authentication and rate-limit retries can reuse a request whose body was consumed.
	req.Body.Close()
	body, err := req.GetBody()
	if err != nil {
		return nil, fmt.Errorf("reopen registry request body: %w", err)
	}

	cloned := req.Clone(req.Context())
	cloned.Body = body
	return t.underlying.RoundTrip(cloned)
}
