package parallel

import (
	"github.com/werf/common-go/pkg/util"
)

type bufferedWorker struct {
	ID  int
	buf *util.GoroutineSafeBuffer

	doneCh chan struct{}
}

func (w *bufferedWorker) MarkDone() {
	w.doneCh <- struct{}{}
}

func (w *bufferedWorker) Buffer() *util.GoroutineSafeBuffer {
	return w.buf
}

func (w *bufferedWorker) IsDone() bool {
	return len(w.doneCh) > 0
}

func newBufferedWorker(id int) *bufferedWorker {
	return &bufferedWorker{
		ID:     id,
		buf:    util.NewGoroutineSafeBuffer(),
		doneCh: make(chan struct{}, 1),
	}
}
