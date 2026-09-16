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
// the task's own log lines. Outside a task, relayed writes are dropped.
type Worker struct {
	ID int

	mu      sync.Mutex
	current *TaskOutput
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

// Output returns the TaskOutput of the task the worker runs right now, nil
// outside a task.
func (w *Worker) Output() *TaskOutput {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.current
}

// beginTask creates the buffer for the task about to run.
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

// bindTaskStreams points the docker cli relay at the task logger's streams.
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
