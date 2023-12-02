package transport

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/werf/logboek"
	parallelConstant "github.com/werf/werf/pkg/util/parallel/constant"
)

type RetryAfter struct {
	underlying http.RoundTripper
}

func NewRetryAfter(underlying http.RoundTripper) http.RoundTripper {
	return &RetryAfter{underlying: underlying}
}

func (t *RetryAfter) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.underlying.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if retryAfterHeader := resp.Header.Get("Retry-After"); retryAfterHeader != "" {
		if seconds, err := strconv.Atoi(retryAfterHeader); err == nil {
			sleepSeconds := rand.Intn(15) + 5

			ctx := req.Context()
			workerId := ctx.Value(parallelConstant.CtxBackgroundTaskIDKey)
			if workerId != nil {
				logboek.Context(ctx).Warn().LogF(
					"WARNING: Rate limit error occurred. Waiting for %d before retrying request... (worker %d).\nThe --parallel ($WERF_PARALLEL) and --parallel-tasks-limit ($WERF_PARALLEL_TASKS_LIMIT) options can be used to regulate parallel tasks.\n",
					sleepSeconds,
					workerId.(int),
				)
				logboek.Context(ctx).Warn().LogLn()
			} else {
				logboek.Context(ctx).Warn().LogF(
					"WARNING: Rate limit error occurred. Waiting for %d before retrying request...\n",
					sleepSeconds,
				)
			}

			time.Sleep(time.Second * time.Duration(seconds))
			return t.RoundTrip(req)
		}
	}

	return resp, nil
}
