package parallel

import (
	"fmt"
	"os"
	"sync/atomic"
)

// Worker
// This is a worker for concurrent writing and reading logs.
// It uses *os.File under the hood.
//
// To be concurrently safe, writer and reader rely on the same file object:
// - the writer appends data only;
// - the reader reads already appended data (or nothing).
// Because of that, no race condition happens while accessing file data.
type Worker struct {
	ID int

	file *os.File

	failed     atomic.Bool
	halfClosed atomic.Bool

	writeOffset atomic.Int64
	readOffset  atomic.Int64
}

// Write implements io.Writer.
// It appends to file and accumulates total write offset.
func (w *Worker) Write(p []byte) (int, error) {
	if w.halfClosed.Load() {
		// TODO (zaytsev): a workaround for non-blocking writing to be able to close underlying file safely
		// fmt.Printf("worker %d is half closed but tries to write: %s\n", w.ID, p)
		return len(p), nil

		// TODO (zaytsev): uncomment after fixing of parallel logic (a task must be executed in blocking mode to correctly handle deferred calls)
		// return 0, fmt.Errorf("worker is half closed but tries to write: %s", p)
	}

	offset, err := w.file.Write(p)
	w.writeOffset.Add(int64(offset))
	return offset, err
}

// Read implements io.Reader.
// It reads a file and accumulates total read offset.
// It resumes reading from "total read offset" and reads until EOF, where EOF is handled with os.File.
func (w *Worker) Read(p []byte) (int, error) {
	offset, err := w.file.ReadAt(p, w.readOffset.Load())
	w.readOffset.Add(int64(offset))
	return offset, err
}

// HalfClose closes writing or returns error if already half-closed
func (w *Worker) HalfClose() error {
	if w.halfClosed.Load() {
		return fmt.Errorf("worker %d is already half closed", w.ID)
	}
	w.halfClosed.Store(true)
	return nil
}

// Fail marks worker as failed
func (w *Worker) Fail() {
	w.failed.Store(true)
}

// Failed returns true if worker failed
func (w *Worker) Failed() bool {
	return w.failed.Load()
}

// Readable returns true if worker is readable
func (w *Worker) Readable() bool {
	if !w.halfClosed.Load() {
		return true
	}
	return w.readOffset.Load() < w.writeOffset.Load()
}

// Close implements io.Closer closing tmp file.
// It ensures that worker is half closed
func (w *Worker) Close() error {
	w.halfClosed.Store(true)

	if err := w.file.Close(); err != nil {
		return fmt.Errorf("failed to close tmp file for worker %d: %w", w.ID, err)
	}
	return nil
}

// Cleanup removes tmp file
func (w *Worker) Cleanup() error {
	if !w.halfClosed.Load() {
		return fmt.Errorf("worker %d is not half closed yet", w.ID)
	}

	if err := os.Remove(w.file.Name()); err != nil {
		return fmt.Errorf("failed to remove tmp file for worker %d: %w", w.ID, err)
	}
	return nil
}

func NewWorker(id int) (*Worker, error) {
	file, err := os.CreateTemp("", fmt.Sprintf("parallel-worker-%d-%d-*.log", os.Getpid(), id))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for worker %d: %w", id, err)
	}

	return &Worker{
		ID:   id,
		file: file,
	}, nil
}
