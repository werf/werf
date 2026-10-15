package storage

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/logging"
)

const metaMigrationProgressInterval = 10 * time.Second

type migrationLogWriter struct {
	writer io.Writer
	mu     *sync.Mutex
}

var _ io.Writer = (*migrationLogWriter)(nil)

func (w *migrationLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(p)
}

func runMetaMigrationPhase(ctx context.Context, name string, total, workers int, interval time.Duration, task func(context.Context, int) error) error {
	var mu sync.Mutex
	logger := logboek.Context(ctx)
	out := &migrationLogWriter{writer: logger.OutStream(), mu: &mu}
	errOut := &migrationLogWriter{writer: logger.ErrStream(), mu: &mu}
	// Workers and heartbeat must not share logboek's mutable stream state.
	progressLogger := logging.NewSubLogger(ctx, out, errOut)
	var completed atomic.Int64
	started := time.Now()
	report := func(status string) {
		progressLogger.Default().LogF("Metadata migration: %s: %d/%d completed, elapsed %s (%s)\n", name, completed.Load(), total, time.Since(started).Round(time.Second), status)
	}
	group, groupCtx := errgroup.WithContext(ctx)
	numberOfWorkers := min(total, max(1, workers))
	workerContexts := make([]context.Context, numberOfWorkers)
	for worker := range workerContexts {
		workerContexts[worker] = logboek.NewContext(groupCtx, logging.NewSubLogger(ctx, out, errOut))
	}
	report("running")
	result := make(chan error, 1)
	go func() {
		for worker, workerCtx := range workerContexts {
			group.Go(func() error {
				for index := worker; index < total; index += numberOfWorkers {
					if err := workerCtx.Err(); err != nil {
						return err
					}
					if err := task(workerCtx, index); err != nil {
						return fmt.Errorf("task %d: %w", index, err)
					}
					completed.Add(1)
				}
				return nil
			})
		}
		result <- group.Wait()
	}()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			report("running")
		case err := <-result:
			if err != nil {
				report("interrupted")
				return fmt.Errorf("%s: %w", name, err)
			}
			report("done")
			return nil
		}
	}
}
