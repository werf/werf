package parallel

import (
	"fmt"
	"os"
	"sync/atomic"
	"unicode/utf8"

	"github.com/werf/werf/v2/pkg/tmp_manager"
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
	readOffset  atomic.Int64
	writeOffset atomic.Int64

	halfClosed atomic.Bool

	file *os.File

	ID int
}

// Write implements io.Writer.
// It appends to file and accumulates total write offset.
func (w *Worker) Write(p []byte) (int, error) {
	if w.halfClosed.Load() {
		return 0, fmt.Errorf("worker is half closed but tries to write: %s", p)
	}

	offset, err := w.file.Write(p)
	w.writeOffset.Add(int64(offset))
	return offset, err
}

// Read implements io.Reader.
// It reads a file and accumulates total read offset.
// It resumes reading from "total read offset" and reads until EOF, where EOF is handled with os.File.
//
// A trailing incomplete UTF-8 sequence is held back and returned on the next
// call as long as more bytes are still to come (either the worker is still
// writing, or half-closed but not yet fully drained). Without this, a
// fixed-size read can land its boundary in the middle of a multi-byte rune
// (e.g. a box-drawing character used for log prefixes), and the downstream
// logger converts each half independently into a replacement character,
// producing visible mojibake in the terminal.
func (w *Worker) Read(p []byte) (int, error) {
	readOffset := w.readOffset.Load()
	n, err := w.file.ReadAt(p, readOffset)

	atEnd := w.halfClosed.Load() && readOffset+int64(n) >= w.writeOffset.Load()
	if !atEnd {
		if complete := completeUTF8Len(p[:n]); complete > 0 && complete < n {
			n = complete
			err = nil
		}
	}

	w.readOffset.Add(int64(n))
	return n, err
}

// completeUTF8Len returns the length of the longest prefix of b that does not
// end with a truncated multi-byte UTF-8 sequence.
func completeUTF8Len(b []byte) int {
	n := len(b)

	for i := 1; i < utf8.UTFMax && i <= n; i++ {
		c := b[n-i]
		if utf8.RuneStart(c) {
			if utf8SequenceLen(c) > i {
				return n - i
			}
			break
		}
	}

	return n
}

// utf8SequenceLen returns the expected total byte length of the UTF-8
// sequence starting with lead byte c.
func utf8SequenceLen(c byte) int {
	switch {
	case c&0x80 == 0x00:
		return 1
	case c&0xE0 == 0xC0:
		return 2
	case c&0xF0 == 0xE0:
		return 3
	case c&0xF8 == 0xF0:
		return 4
	default:
		return 1
	}
}

// HalfClose closes writing or returns error if already half-closed
func (w *Worker) HalfClose() error {
	if w.halfClosed.Load() {
		return fmt.Errorf("worker %d is already half closed", w.ID)
	}
	w.halfClosed.Store(true)
	return nil
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
	file, err := tmp_manager.TempFile(fmt.Sprintf("parallel-worker-%d-%d-*.log", os.Getpid(), id))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for worker %d: %w", id, err)
	}

	return &Worker{
		ID:   id,
		file: file,
	}, nil
}
