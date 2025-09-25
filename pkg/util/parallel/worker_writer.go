package parallel

import (
	"io"
	"os"
)

type parallelWorkerWriter struct {
	file   *os.File
	doneCh chan struct{}
}

func (w *parallelWorkerWriter) Stream() io.Writer {
	return w.file
}

func (w *parallelWorkerWriter) Close() {
	// unblocking write
	select {
	case w.doneCh <- struct{}{}:
	default:
	}
}

func (w *parallelWorkerWriter) Done() <-chan struct{} {
	return w.doneCh
}

func newParallelWorkerWriter(f *os.File) *parallelWorkerWriter {
	return &parallelWorkerWriter{
		file:   f,
		doneCh: make(chan struct{}, 1),
	}
}
