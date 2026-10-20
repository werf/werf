package transport

import (
	"io"
	"net/http"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

var _ http.RoundTripper = roundTripperFunc(nil)

func (roundTrip roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return roundTrip(req)
}

type trackedResponseBody struct {
	io.Reader
	closed bool
}

var _ io.ReadCloser = (*trackedResponseBody)(nil)

func (body *trackedResponseBody) Close() error {
	body.closed = true
	return nil
}
