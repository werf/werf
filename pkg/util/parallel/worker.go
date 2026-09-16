package parallel

import (
	"errors"
	"fmt"
	"io"
	"sync"
)

// Worker is the io.Writer a worker's sub-logger is bound to for the whole
// run. It forwards every write to the TaskOutput of the task the worker is
// running right now, so the logger and the docker cli attached to the
// worker context are created once while each task still gets its own
// buffer. Outside a task, writes are dropped.
type Worker struct {
	ID int

	mu      sync.Mutex
	current *TaskOutput
	outputs []*TaskOutput
}

var _ io.Writer = (*Worker)(nil)

func NewWorker(id int) *Worker {
	return &Worker{ID: id}
}

// Write implements io.Writer.
func (w *Worker) Write(p []byte) (int, error) {
	out := w.Output()
	if out == nil {
		return len(p), nil
	}

	return out.Write(p)
}

// Output returns the TaskOutput the worker writes to right now, nil outside
// a task.
func (w *Worker) Output() *TaskOutput {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.current
}

// beginTask switches writes to a fresh buffer for the task about to run.
func (w *Worker) beginTask() (*TaskOutput, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	out, err := NewTaskOutput(w.ID, len(w.outputs))
	if err != nil {
		return nil, err
	}

	w.current = out
	w.outputs = append(w.outputs, out)
	return out, nil
}

// endTask stops accepting output of the current task.
func (w *Worker) endTask() {
	if out := w.Output(); out != nil {
		out.HalfClose()
	}
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
