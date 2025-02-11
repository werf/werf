package transport

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
)

const retryAfterHeaderKey = "Retry-After"

var ErrRetryAfterHeaderNotPresent = errors.New("retry after header not present")

// backoffHttpRetryAfterHandler Handles http "Retry-After" header for 301, 429, 503 statuses
//
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After
func backoffHttpRetryAfterHandler(resp *http.Response) (*http.Response, error) {
	specHeaders := []int{
		http.StatusMovedPermanently,
		http.StatusTooManyRequests,
		http.StatusServiceUnavailable,
	}

	if !slices.Contains(specHeaders, resp.StatusCode) {
		return resp, nil
	}

	if len(resp.Header.Values(retryAfterHeaderKey)) == 0 {
		// header not present, retry with strategy configured before
		return nil, ErrRetryAfterHeaderNotPresent
	}

	// date or seconds
	headerValue := resp.Header.Get(retryAfterHeaderKey)

// ------------------
// Try parsing as <http-date> (RFC1123 format).
// Example: 'Tue, 29 Oct 2024 16:56:32 GMT'.
// ------------------
if retryAfterTime, err := time.Parse(time.RFC1123, headerValue); err == nil {
	seconds := time.Until(retryAfterTime).Seconds()
	return nil, backoff.RetryAfter(int(math.Max(0, seconds)))
}

// ------------------
// Try parsing as <delay-seconds>.
// IMPORTANT: Constraint with uint8 (bitSize) → max=255 (4 min 15 sec).
seconds, err := strconv.ParseUint(headerValue, 10, 8)
// ------------------
if err != nil {
	return nil, backoff.Permanent(err) // Consider parsing error as permanent.
}

return nil, backoff.RetryAfter(int(seconds))
}
