package parallel

import (
	"fmt"
	"io"
	"os"
)

type bufferedWorker struct {
	ID int

	outStream           *os.File
	readerOffset        int
	readerLastReadBytes int

	doneCh chan struct{}
}

func (w *bufferedWorker) MarkWritingDone() {
	// unblocking write
	select {
	case w.doneCh <- struct{}{}:
	default:
	}
}

func (w *bufferedWorker) OutStream() *os.File {
	return w.outStream
}

func (w *bufferedWorker) ReadFromStream() ([]byte, error) {
	if _, err := w.outStream.Seek(int64(w.readerOffset), 0); err != nil {
		return nil, fmt.Errorf("failed to seek to offset %d: %w", w.readerOffset, err)
	}

	data := make([]byte, 1024)

	var err error
	w.readerLastReadBytes, err = w.outStream.Read(data)

	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	w.readerOffset += w.readerLastReadBytes

	return data[:w.readerLastReadBytes], nil
}

func (w *bufferedWorker) Scan() bool {
	if len(w.doneCh) > 0 && w.readerLastReadBytes == 0 {
		return false
	}
	return true
}

func (w *bufferedWorker) Close() error {
	return w.outStream.Close()
}

func newBufferedWorker(id int) (*bufferedWorker, error) {
	pid := os.Getpid()

	file, err := os.CreateTemp("", fmt.Sprintf("parallel-worker-%d-%d", pid, id))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for worker %d: %w", id, err)
	}

	return &bufferedWorker{
		ID:     id,
		doneCh: make(chan struct{}, 1),

		outStream:           file,
		readerOffset:        0,
		readerLastReadBytes: 0,
	}, nil
}
