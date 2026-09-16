package parallel

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"unicode/utf8"

	"github.com/werf/werf/v2/pkg/tmp_manager"
)

// TaskOutput buffers the log of a single task in a temp file so it can be
// written by the task and read by the Printer concurrently.
//
// To be concurrently safe, writer and reader rely on the same file object:
// - the writer appends data only;
// - the reader reads already appended data (or nothing).
// Because of that, no race condition happens while accessing file data.
type TaskOutput struct {
	readOffset  atomic.Int64
	writeOffset atomic.Int64

	mutex      sync.Mutex
	halfClosed atomic.Bool

	file *os.File
}

// Write implements io.Writer.
// It appends to file and accumulates total write offset.
func (o *TaskOutput) Write(p []byte) (int, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.halfClosed.Load() {
		return len(p), nil
	}

	offset, err := o.file.Write(p)
	o.writeOffset.Add(int64(offset))
	return offset, err
}

// Read implements io.Reader.
// It reads a file and accumulates total read offset.
// It resumes reading from "total read offset" and reads until EOF, where EOF is handled with os.File.
//
// A trailing incomplete UTF-8 sequence is held back and returned on the next
// call as long as more bytes are still to come (either the task is still
// writing, or half-closed but not yet fully drained). Without this, a
// fixed-size read can land its boundary in the middle of a multi-byte rune
// (e.g. a box-drawing character used for log prefixes), and the downstream
// logger converts each half independently into a replacement character,
// producing visible mojibake in the terminal.
func (o *TaskOutput) Read(p []byte) (int, error) {
	readOffset := o.readOffset.Load()
	n, err := o.file.ReadAt(p, readOffset)

	atEnd := o.halfClosed.Load() && readOffset+int64(n) >= o.writeOffset.Load()
	if !atEnd {
		if complete := completeUTF8Len(p[:n]); complete > 0 && complete < n {
			n = complete
			err = nil
		}
	}

	o.readOffset.Add(int64(n))
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

// HalfClose stops accepting writes; later writes are silently dropped.
// Calling it on an already half-closed output is a no-op.
func (o *TaskOutput) HalfClose() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.halfClosed.Store(true)
}

// Readable returns true while there is (or may still come) something to read.
func (o *TaskOutput) Readable() bool {
	if !o.halfClosed.Load() {
		return true
	}
	return o.readOffset.Load() < o.writeOffset.Load()
}

// Close implements io.Closer closing tmp file.
// It ensures that the output is half closed.
func (o *TaskOutput) Close() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.halfClosed.Store(true)

	if err := o.file.Close(); err != nil {
		return fmt.Errorf("close tmp file %q: %w", o.file.Name(), err)
	}
	return nil
}

// Cleanup removes tmp file
func (o *TaskOutput) Cleanup() error {
	if !o.halfClosed.Load() {
		return fmt.Errorf("task output %q is not half closed yet", o.file.Name())
	}

	if err := os.Remove(o.file.Name()); err != nil {
		return fmt.Errorf("remove tmp file %q: %w", o.file.Name(), err)
	}
	return nil
}

func NewTaskOutput(workerID, taskSeq int) (*TaskOutput, error) {
	file, err := tmp_manager.TempFile(fmt.Sprintf("parallel-worker-%d-%d-%d-*.log", os.Getpid(), workerID, taskSeq))
	if err != nil {
		return nil, fmt.Errorf("create temp file for worker %d task %d: %w", workerID, taskSeq, err)
	}

	return &TaskOutput{
		file: file,
	}, nil
}
