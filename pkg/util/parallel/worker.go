package parallel

import (
	"fmt"
	"os"
	"time"
)

type parallelWorker struct {
	ID int

	file *os.File

	Writer *parallelWorkerWriter
	Reader *parallelWorkerReader
}

func (w *parallelWorker) Close() error {
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("failed to close temp file for worker %d: %w", w.ID, err)
	}
	if err := os.Remove(w.file.Name()); err != nil {
		return fmt.Errorf("failed to remove temp file for worker %d: %w", w.ID, err)
	}
	return nil
}

func newParallelWorker(id int) (*parallelWorker, error) {
	pid := os.Getpid()

	file, err := os.CreateTemp("", fmt.Sprintf("parallel-worker-%d-%d-%d", pid, id, time.Now().UnixMilli()))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for worker %d: %w", id, err)
	}

	return &parallelWorker{
		ID:     id,
		file:   file,
		Writer: newParallelWorkerWriter(file),
		Reader: newParallelWorkerReader(file),
	}, nil
}
