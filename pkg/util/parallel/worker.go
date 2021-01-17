package parallel

import (
	"bytes"
)

type bufWorker struct {
	buf *bytes.Buffer
}

func (w *bufWorker) IsLiveWorker() bool {
	return false
}

func (w *bufWorker) TaskResult(err error) interface{} {
	taskResult := &bufWorkerTaskResult{
		buf:  w.buf,
		err:  err,
		data: w.buf.Bytes(),
	}

	w.buf.Reset()

	return taskResult
}

type bufWorkerTaskResult struct {
	buf  *bytes.Buffer
	data []byte
	err  error
}

type liveWorker struct{}

func (w *liveWorker) IsLiveWorker() bool {
	return true
}

func (w *liveWorker) TaskResult(err error) interface{} {
	return &lifeWorkerTaskResult{err: err}
}

type lifeWorkerTaskResult struct {
	err error
}

type worker interface {
	TaskResult(err error) interface{}
	IsLiveWorker() bool
}
