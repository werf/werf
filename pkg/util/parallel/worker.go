package parallel

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/samber/lo"
)

// Worker owns the TaskOutputs of the tasks it ran and relays the output of
// the worker-level docker cli to the logger of the task running right now.
// The cli is created once per worker (a client per task would leak a
// connection pool per task) and writes only synchronously from inside the
// task that invoked it, so relaying to the current task's logger streams is
// exact and its output gets the same indentation and block boundaries as
// the task's own log lines. Between tasks relayed writes are dropped; a
// write from a goroutine that outlived its task lands in the block of
// whatever task the worker runs at that moment.
type Worker struct {
	ID int

	mu      sync.Mutex
	current *TaskOutput
	failed  *TaskOutput
	outputs []*TaskOutput
	taskOut io.Writer
	taskErr io.Writer
}

func NewWorker(id int) *Worker {
	return &Worker{ID: id}
}

// OutStream and ErrStream are what the worker's docker cli is bound to.
func (w *Worker) OutStream() io.Writer { return streamRelay{worker: w, err: false} }
func (w *Worker) ErrStream() io.Writer { return streamRelay{worker: w, err: true} }

type streamRelay struct {
	worker *Worker
	err    bool
}

var _ io.Writer = streamRelay{}

func (r streamRelay) Write(p []byte) (int, error) {
	r.worker.mu.Lock()
	target := lo.Ternary(r.err, r.worker.taskErr, r.worker.taskOut)
	r.worker.mu.Unlock()

	if target == nil {
		return len(p), nil
	}

	return target.Write(p)
}

// failTask records the output of a task that returned an error, so the
// printer can highlight exactly that block. An error raised outside a task
// (picking the next one) leaves it unset, and the printing queue is then
// left alone.
func (w *Worker) failTask(out *TaskOutput) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.failed = out
}

func (w *Worker) failedOutput() *TaskOutput {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.failed
}

func (w *Worker) beginTask() *TaskOutput {
	w.mu.Lock()
	defer w.mu.Unlock()

	out := NewTaskOutput(w.ID, len(w.outputs))
	w.current = out
	w.outputs = append(w.outputs, out)
	return out
}

func (w *Worker) bindTaskStreams(outStream, errStream io.Writer) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.taskOut = outStream
	w.taskErr = errStream
}

// endTask detaches the docker cli relay and stops accepting output of the
// current task.
func (w *Worker) endTask() error {
	w.mu.Lock()
	w.taskOut = nil
	w.taskErr = nil
	out := w.current
	w.current = nil
	w.mu.Unlock()

	if out == nil {
		return nil
	}

	return out.HalfClose()
}

// Close closes the tmp files of every task the worker ran.
func (w *Worker) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	var errs []error
	for _, out := range w.outputs {
		if err := out.Close(); err != nil {
			errs = append(errs, fmt.Errorf("worker %d: %w", w.ID, err))
		}
	}
	return errors.Join(errs...)
}

// Cleanup removes the tmp files of every task the worker ran.
func (w *Worker) Cleanup() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	var errs []error
	for _, out := range w.outputs {
		if err := out.Cleanup(); err != nil {
			errs = append(errs, fmt.Errorf("worker %d: %w", w.ID, err))
		}
	}
	return errors.Join(errs...)
}
