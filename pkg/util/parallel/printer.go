package parallel

import (
	"context"
	"fmt"
	"io"
	"slices"
	"sync"
	"time"

	"github.com/werf/logboek"
)

// Printer renders task output as coherent, uninterrupted per-task blocks in
// the order the tasks started: it streams the oldest unfinished task live
// and fully drains it before moving to the next, rather than interleaving
// concurrent tasks' output line by line. Every task is enqueued the moment
// it starts, so its position doubles as the task's start-order index.
//
// Because the head of the queue is always the oldest running task, the live
// log never goes quiet while something is still building — as soon as the
// head finishes, the next task in start order takes over, already partially
// buffered.
type Printer struct {
	mu     sync.Mutex
	queue  []*TaskOutput
	cursor int
	closed bool
}

func NewPrinter() *Printer {
	return &Printer{}
}

// Enqueue appends a started task to the printing queue and returns its
// start-order index.
func (p *Printer) Enqueue(out *TaskOutput) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.queue = append(p.queue, out)
	return len(p.queue) - 1
}

// Close tells the Printer no more tasks will be enqueued, so Print returns
// once the queue is drained.
func (p *Printer) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
}

// FailFast reorders the queue after a task failed: a failed task that is not
// being printed yet moves to the end, so its error is the last thing the
// user sees; a failed task that is being printed right now truncates the
// queue behind it — the tasks queued after it were canceled and their
// partial output is noise.
func (p *Printer) FailFast(failed *TaskOutput) {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := slices.Index(p.queue, failed)
	switch {
	case idx < p.cursor:
		return
	case idx == p.cursor:
		p.queue = p.queue[:idx+1]
	default:
		p.queue = append(slices.Delete(p.queue, idx, idx+1), failed)
	}
}

// Print streams the queue in order and returns once the Printer is closed
// and drained, or ctx is done. Calling it again resumes from where the
// previous call stopped.
func (p *Printer) Print(ctx context.Context) error {
	for {
		out, ok, done := p.head()
		if done {
			return nil
		}

		if !ok {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}

		if err := printTaskOutput(ctx, out); err != nil {
			return err
		}

		p.mu.Lock()
		p.cursor++
		p.mu.Unlock()
	}
}

func (p *Printer) head() (out *TaskOutput, ok, done bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cursor < len(p.queue) {
		return p.queue[p.cursor], true, false
	}

	return nil, false, p.closed
}

func printTaskOutput(ctx context.Context, out *TaskOutput) error {
	var offset int64
	var err error

	buf := make([]byte, 1024)

	for out.Readable() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err = logboek.Context(ctx).Streams().DoErrorWithoutIndent(func() error {
				offset, err = io.CopyBuffer(logboek.Context(ctx).OutStream(), out, buf)
				return err
			}); err != nil {
				return fmt.Errorf("copy task output: %w", err)
			}

			clear(buf)

			if offset == 0 {
				time.Sleep(time.Millisecond * 100)
			}
		}
	}

	logboek.Context(ctx).LogOptionalLn()

	return ctx.Err()
}
