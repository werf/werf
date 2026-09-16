package parallel

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"unicode/utf8"

	"github.com/werf/werf/v2/pkg/tmp_manager"
)

// TaskOutput buffers the log of a single task in a temp file so it can be
// written by the task and read by the Printer concurrently.
//
// The writer appends only and the reader reads already appended data (or
// nothing), so they never race on file content. The writer descriptor is
// closed at HalfClose and the reader descriptor is opened on first Read and
// closed once everything is drained: a finished task waiting to be printed
// holds no descriptor, which keeps the count of open files bounded by the
// number of workers rather than the number of tasks.
type TaskOutput struct {
	mutex sync.Mutex

	path        string
	writer      *os.File // nil once half-closed
	reader      *os.File // open only while being drained
	readOffset  int64
	writeOffset int64
	lastByte    byte
}

// Write implements io.Writer.
// It appends to file and accumulates total write offset.
func (o *TaskOutput) Write(p []byte) (int, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.writer == nil {
		return len(p), nil
	}

	n, err := o.writer.Write(p)
	o.writeOffset += int64(n)
	if n > 0 {
		o.lastByte = p[n-1]
	}
	return n, err
}

// Read implements io.Reader.
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
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.reader == nil {
		reader, err := os.Open(o.path)
		if err != nil {
			return 0, fmt.Errorf("open task output for reading: %w", err)
		}
		o.reader = reader
	}

	n, err := o.reader.ReadAt(p, o.readOffset)

	atEnd := o.writer == nil && o.readOffset+int64(n) >= o.writeOffset
	if !atEnd {
		if complete := completeUTF8Len(p[:n]); complete > 0 && complete < n {
			n = complete
			err = nil
		}
	}

	o.readOffset += int64(n)

	if closeErr := o.releaseDrainedReader(); closeErr != nil {
		err = errors.Join(err, closeErr)
	}

	return n, err
}

// releaseDrainedReader closes the reader once nothing more can come: the
// writer is gone and every byte has been read. Whichever of Read or
// HalfClose completes the drain triggers it. Caller holds o.mutex.
func (o *TaskOutput) releaseDrainedReader() error {
	if o.reader == nil || o.writer != nil || o.readOffset < o.writeOffset {
		return nil
	}

	err := o.reader.Close()
	o.reader = nil
	if err != nil {
		return fmt.Errorf("close task output reader %q: %w", o.path, err)
	}
	return nil
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

// HalfClose stops accepting writes and releases the writer descriptor;
// later writes are silently dropped. Calling it again is a no-op.
//
// If the last byte written is not a newline, one is appended first, under
// the same lock that stops further writes: the block always ends on a line
// boundary, however late the last write came in, so the printer can never
// glue the next block onto it.
func (o *TaskOutput) HalfClose() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.writer == nil {
		return nil
	}

	var errs []error
	if o.writeOffset > 0 && o.lastByte != '\n' {
		n, err := o.writer.Write([]byte{'\n'})
		o.writeOffset += int64(n)
		if err != nil {
			errs = append(errs, fmt.Errorf("terminate task output %q: %w", o.path, err))
		}
	}

	if err := o.writer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close task output writer %q: %w", o.path, err))
	}
	o.writer = nil

	if err := o.releaseDrainedReader(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// Readable returns true while there is (or may still come) something to read.
func (o *TaskOutput) Readable() bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return o.writer != nil || o.readOffset < o.writeOffset
}

// Close implements io.Closer: it half-closes the output and releases the
// reader descriptor if a drain was interrupted.
func (o *TaskOutput) Close() error {
	if err := o.HalfClose(); err != nil {
		return err
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.reader == nil {
		return nil
	}

	err := o.reader.Close()
	o.reader = nil
	if err != nil {
		return fmt.Errorf("close task output reader %q: %w", o.path, err)
	}
	return nil
}

// Cleanup removes tmp file
func (o *TaskOutput) Cleanup() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.writer != nil {
		return fmt.Errorf("task output %q is not half closed yet", o.path)
	}

	if err := os.Remove(o.path); err != nil {
		return fmt.Errorf("remove tmp file %q: %w", o.path, err)
	}
	return nil
}

func NewTaskOutput(workerID, taskSeq int) (*TaskOutput, error) {
	file, err := tmp_manager.TempFile(fmt.Sprintf("parallel-worker-%d-%d-%d-*.log", os.Getpid(), workerID, taskSeq))
	if err != nil {
		return nil, fmt.Errorf("create temp file for worker %d task %d: %w", workerID, taskSeq, err)
	}

	return &TaskOutput{
		path:   file.Name(),
		writer: file,
	}, nil
}
