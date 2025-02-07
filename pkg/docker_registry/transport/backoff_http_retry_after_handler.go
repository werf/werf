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
	// Is it <http-date>? Sample: 'Tue, 29 Oct 2024 16:56:32 GMT'
	// ------------------
	if strings.Count(headerValue, " ") == 5 {
		retryAfterTime, err := time.Parse(time.RFC1123, headerValue)
		if err != nil {
			return nil, backoff.Permanent(err)
		}

		seconds := retryAfterTime.Sub(time.Now()).Seconds()
		return nil, backoff.RetryAfter(int(seconds))
	}

	// ------------------
	// It is <delay-seconds>.
	// ------------------
	// IMPORTANT: This is constraint with uint8 (bitSize).
	// It means max=255, what is 255/60 = 4.25 min or 4 min 15 sec
	seconds, err := strconv.ParseUint(headerValue, 10, 8)
	// Consider str_conv err as permanent one
	if err != nil {
		return nil, backoff.Permanent(err)
	}

	return nil, backoff.RetryAfter(int(seconds))
}
