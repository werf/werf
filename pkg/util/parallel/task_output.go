package parallel

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"unicode/utf8"

	"github.com/werf/werf/v2/pkg/tmp_manager"
)

// TaskOutput buffers the log of a single task in a temp file so it can be
// written by the task and read by the Printer concurrently.
//
// The file is created by the first write, so a task that logs nothing never
// touches the disk: most tasks of a cleanup run are silent, and a buffer per
// task allocated up front would put an inode per task in the tmp dir for as
// long as the printing queue holds them back.
//
// The writer appends only and the reader reads already appended data (or
// nothing), so they never race on file content. The writer descriptor is
// closed at HalfClose and the reader descriptor is opened on first Read and
// closed once everything is drained: a finished task waiting to be printed
// holds no descriptor, which keeps the count of open files bounded by the
// number of workers rather than the number of tasks.
//
// Any number of goroutines may write; only one may read, and Close may not
// run while a Read is in flight (the Printer is the sole reader and is
// stopped before the outputs are closed).
type TaskOutput struct {
	mutex sync.Mutex

	workerID int
	taskSeq  int

	path        string
	writer      *os.File // nil until the first write and once half-closed
	createErr   error
	halfClosed  bool
	reader      *os.File // open only while being drained
	reading     bool
	readOffset  int64
	writeOffset int64
	lastByte    byte
	removed     bool
}

// Write implements io.Writer.
// It appends to file and accumulates total write offset.
//
// Failing to create the buffer costs the task's log, not the task: logboek
// discards whatever a log write returns, so the error is kept for HalfClose
// to report instead.
func (o *TaskOutput) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.halfClosed || o.createErr != nil {
		return len(p), nil
	}

	if o.writer == nil {
		file, err := tmp_manager.TempFile(fmt.Sprintf("parallel-worker-%d-%d-%d-*.log", os.Getpid(), o.workerID, o.taskSeq))
		if err != nil {
			o.createErr = fmt.Errorf("create temp file for worker %d task %d: %w", o.workerID, o.taskSeq, err)
			return len(p), nil
		}
		o.path = file.Name()
		o.writer = file
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
// producing visible mojibake in the terminal. A read that finds nothing but
// such a fragment returns io.EOF, so the reader comes back for it once the
// rest has been written.
func (o *TaskOutput) Read(p []byte) (int, error) {
	o.mutex.Lock()

	if o.path == "" || (o.halfClosed && o.readOffset >= o.writeOffset) {
		o.mutex.Unlock()
		return 0, io.EOF
	}

	if o.reader == nil {
		reader, err := os.Open(o.path)
		if err != nil {
			o.mutex.Unlock()
			return 0, fmt.Errorf("open task output for reading: %w", err)
		}
		o.reader = reader
	}

	// The read itself runs unlocked so the task keeps logging while the
	// printer waits on the disk; o.reading holds the descriptor open until
	// it is done.
	reader, readOffset := o.reader, o.readOffset
	o.reading = true
	o.mutex.Unlock()

	n, err := reader.ReadAt(p, readOffset)

	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.reading = false

	atEnd := o.halfClosed && o.readOffset+int64(n) >= o.writeOffset
	if !atEnd {
		complete := completeUTF8Len(p[:n])
		if complete == 0 && n == len(p) {
			// p cannot hold a whole rune, so the fragment has to go through
			// or the drain never advances.
			complete = n
		}
		if complete < n {
			n = complete
			// A read left with nothing but the fragment reports EOF: that
			// parks the printer until more is written, where 0 and no error
			// would spin it.
			if n == 0 {
				err = io.EOF
			} else {
				err = nil
			}
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
	if o.reader == nil || o.reading || !o.halfClosed || o.readOffset < o.writeOffset {
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
// later writes are silently dropped. Calling it again is a no-op. It reports
// a buffer that could not be created, and with it the loss of the task's log.
//
// If the last byte written is not a newline, one is appended first, under
// the same lock that stops further writes: the block always ends on a line
// boundary, however late the last write came in, so the printer can never
// glue the next block onto it.
func (o *TaskOutput) HalfClose() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.halfClosed {
		return nil
	}
	o.halfClosed = true

	var errs []error
	if o.createErr != nil {
		errs = append(errs, o.createErr)
	}

	if o.writer != nil {
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
	}

	if err := o.releaseDrainedReader(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// Readable returns true while there is (or may still come) something to read.
func (o *TaskOutput) Readable() bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return !o.halfClosed || o.readOffset < o.writeOffset
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

// Cleanup removes the tmp file. The Printer calls it as soon as the output
// is drained, so the disk holds only what is not printed yet; the final
// sweep in runWorkers calls it again for whatever the Printer never reached.
// A call after a successful removal is a no-op, while a removal that failed
// is retried by the sweep: the Printer only warns about it, so giving up
// after the first failure would leak the file for the rest of the run.
func (o *TaskOutput) Cleanup() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.removed {
		return nil
	}

	if !o.halfClosed {
		return fmt.Errorf("task output of worker %d task %d is not half closed yet", o.workerID, o.taskSeq)
	}
	if o.path != "" {
		if err := os.Remove(o.path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove tmp file %q: %w", o.path, err)
		}
	}

	o.removed = true
	return nil
}

func NewTaskOutput(workerID, taskSeq int) *TaskOutput {
	return &TaskOutput{workerID: workerID, taskSeq: taskSeq}
}
