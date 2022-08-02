package parallel

import "github.com/werf/werf/pkg/util"

type bufWorker struct {
	buf    *util.GoroutineSafeBuffer
	isDone bool
}

func (w *bufWorker) TaskResult(err error) *bufWorkerTaskResult {
	return &bufWorkerTaskResult{
		worker: w,
		err:    err,
	}
}

type bufWorkerTaskResult struct {
	worker *bufWorker
	err    error
}
