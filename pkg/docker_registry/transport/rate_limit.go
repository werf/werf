package transport

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/werf/logboek"
	parallelConstant "github.com/werf/werf/v2/pkg/util/parallel/constant"
)

type RateLimit struct {
	underlying http.RoundTripper
}

func NewRateLimit(underlying http.RoundTripper) http.RoundTripper {
	return &RateLimit{underlying: underlying}
}

type rateLimitError error

func (t *RateLimit) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	operation := func() error {
		var err error
		resp, err = t.underlying.RoundTrip(req)
		if err != nil {
			return backoff.Permanent(err)
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			return nil
		}

		// Ensure response body is closed if retrying.
		defer resp.Body.Close()

		return rateLimitError(errors.New(resp.Status))
	}

	notify := func(err error, duration time.Duration) {
		ctx := req.Context()
		workerId := ctx.Value(parallelConstant.CtxBackgroundTaskIDKey)
		if workerId != nil {
			logboek.Context(ctx).Warn().LogF(
				"WARNING: %s. Retrying in %v... (worker %d).\nThe --parallel ($WERF_PARALLEL) and --parallel-tasks-limit ($WERF_PARALLEL_TASKS_LIMIT) options can be used to regulate parallel tasks.\n",
				err,
				duration,
				workerId.(int),
			)
			logboek.Context(ctx).Warn().LogLn()
		} else {
			logboek.Context(ctx).Warn().LogF(
				"WARNING: %s. Retrying in %v...\n",
				err,
				duration,
			)
		}
	}

	initialInterval := 2 * time.Second
	{
		if err := operation(); err == nil {
			return resp, nil
		}

		if retryAfterHeader := resp.Header.Get("Retry-After"); retryAfterHeader != "" {
			if seconds, err := strconv.Atoi(retryAfterHeader); err == nil {
				initialInterval = time.Duration(seconds) * time.Second
			}
		}
	}

	eb := backoff.NewExponentialBackOff()
	eb.InitialInterval = initialInterval
	eb.MaxElapsedTime = 5 * time.Minute // Maximum time for all retries.
	if err := backoff.RetryNotify(operation, eb, notify); err != nil {
		return nil, err
	}

	return resp, nil
}
