package transport

import (
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v5"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/util/parallel"
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

		return backoffHttpRetryAfterHandler(resp)
	}

	notify := func(err error, duration time.Duration) {
		ctx := req.Context()
		workerId := ctx.Value(parallel.CtxBackgroundTaskIDKey)
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
