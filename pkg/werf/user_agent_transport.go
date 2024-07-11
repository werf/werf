package werf

import (
	"fmt"
	"net/http"
)

var UserAgent = fmt.Sprintf("werf/%s", Version)

type userAgentTransport struct {
	underlying http.RoundTripper
}

func NewUserAgentTransport(underlying http.RoundTripper) http.RoundTripper {
	return &userAgentTransport{underlying: underlying}
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", UserAgent)
	return t.underlying.RoundTrip(req)
}
