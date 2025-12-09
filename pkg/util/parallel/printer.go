package parallel

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/samber/lo"

	"github.com/werf/logboek"
)

type Printer struct {
	workers   []*Worker
	indexes   []int
	cursor    int
	maxCursor int
}

func NewPrinter(workers []*Worker) *Printer {
	return &Printer{
		workers:   workers,
		indexes:   lo.Range(len(workers)),
		cursor:    0,
		maxCursor: len(workers) - 1,
	}
}

func (p *Printer) Cur() int {
	return p.cursor
}

func (p *Printer) Max() int {
	return p.maxCursor
}

func (p *Printer) SetMax(idx int) {
	p.maxCursor = min(idx, len(p.indexes)-1)
}

func (p *Printer) Swap(idx1, idx2 int) {
	val1 := p.indexes[idx1]
	val2 := p.indexes[idx2]
	p.indexes[idx1] = val2
	p.indexes[idx2] = val1
}

func (p *Printer) Print(ctx context.Context) error {
	for ; p.cursor <= p.maxCursor; p.cursor++ {
		idx := p.indexes[p.cursor]

		if err := printWorkerOutput(ctx, p.workers[idx]); err != nil {
			return err
		}
	}

	return nil
}

func printWorkerOutput(ctx context.Context, worker *Worker) error {
	var offset int64
	var err error

	buf := make([]byte, 1024)

	for worker.Readable() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err = logboek.Context(ctx).Streams().DoErrorWithoutIndent(func() error {
				offset, err = io.CopyBuffer(logboek.Context(ctx).OutStream(), worker, buf)
				return err
			}); err != nil {
				return fmt.Errorf("failed to copy output: %w", err)
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
