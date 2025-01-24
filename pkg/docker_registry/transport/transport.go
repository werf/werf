package transport

import (
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v5"

	"github.com/werf/logboek"
	parallelConstant "github.com/werf/werf/v2/pkg/util/parallel/constant"
)

type transport struct {
	underlying http.RoundTripper
}

func NewTransport(underlying http.RoundTripper) http.RoundTripper {
	return &transport{underlying: underlying}
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	operation := func() (*http.Response, error) {
		resp, err := t.underlying.RoundTrip(req)
		if err != nil {
			return nil, backoff.Permanent(err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			// IMPORTANT: This is constraint with uint8 (bitSize).
			// It means max=255, what is 255/60 = 4.25 min or 4 min 15 sec
			seconds, err := strconv.ParseUint(resp.Header.Get("Retry-After"), 10, 8)
			// Consider str_conv err as permanent one
			if err != nil {
				return nil, backoff.Permanent(err)
			}
			return nil, backoff.RetryAfter(int(seconds))
		}

		return resp, nil
	}

	notify := func(err error, duration time.Duration) {
		ctx := req.Context()
		workerId := ctx.Value(parallelConstant.CtxBackgroundTaskIDKey)
		if workerId != nil {
			logboek.Context(ctx).Warn().LogF(
				"WARNING: %s. Retrying in %v... (worker %d)\nThe --parallel ($WERF_PARALLEL) and --parallel-tasks-limit ($WERF_PARALLEL_TASKS_LIMIT) options can be used to regulate parallel tasks.\n",
				err,
				duration,
				workerId.(int),
			)
			logboek.Context(ctx).Warn().LogOptionalLn()
		} else {
			logboek.Context(ctx).Warn().LogF(
				"WARNING: %s. Retrying in %v...\n",
				err,
				duration,
			)
		}
	}

	return backoff.Retry(req.Context(), operation,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxElapsedTime(5*time.Minute), // Maximum time for all retries.
		backoff.WithNotify(notify))
}
