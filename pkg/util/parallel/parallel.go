package parallel

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/logging"
)

type DoTasksOptions struct {
	InitDockerCLIForEachWorker bool
	MaxNumberOfWorkers         int
}

type TaskFunc func(ctx context.Context, taskId int) error

// DoTasks executes a specified number of tasks in parallel using a configurable number of workers.
// Each worker runs a subset of the total tasks, and progress is logged for each task.
//
// Parameters:
//   - ctx: The context used to control the operation and provide cancellation support.
//   - numberOfTasks: The total number of tasks to be executed.
//   - options: A DoTasksOptions struct containing configuration parameters for task execution.
//   - taskFunc: A function that performs a single task. It takes a context and a task ID as input and returns an error if one occurs.
func DoTasks(ctx context.Context, numberOfTasks int, options DoTasksOptions, taskFunc TaskFunc) error {
	logboek.Context(ctx).Debug().LogF("Parallel options: %d (workers) X %d (tasks)\n", options.MaxNumberOfWorkers, numberOfTasks)

	g, groupCtx := errgroup.WithContext(ctx)

	// Determine number of workers and tasks per worker
	numberOfWorkers, numberOfTasksPerWorker := calculateTasksDistribution(numberOfTasks, options.MaxNumberOfWorkers)

	// Create slice of buffered workers
	workers := make([]*parallelWorker, numberOfWorkers)
	// Create channel to signal failed worker
	failedWorkerCh := make(chan int, 1)

	// Close and remove temporary files for workers
	defer func() {
		for _, worker := range workers {
			worker.Close()
		}
	}()

	for i := 0; i < numberOfWorkers; i++ {
		worker, err := newParallelWorker(i)
		if err != nil {
			return fmt.Errorf("failed to create worker %d: %w", i, err)
		}
		workers[worker.ID] = worker

		// Create a new context with a background task ID for each worker
		ctxWithBackgroundTaskID := context.WithValue(groupCtx, CtxBackgroundTaskIDKey, worker.ID)
		workerContext := logboek.NewContext(ctxWithBackgroundTaskID, logging.NewSubLogger(ctxWithBackgroundTaskID, worker.Writer.Stream(), worker.Writer.Stream()))

		if options.InitDockerCLIForEachWorker {
			// TODO: should we always create new docker context for each worker to prevent "context canceled" error?
			if workerContext, err = docker.NewContext(workerContext); err != nil {
				return err
			}
		}

		g.Go(func() error {
			defer worker.Writer.Close()

			for workerTaskId := 0; workerTaskId < numberOfTasksPerWorker[worker.ID]; workerTaskId++ {
				taskId := calculateTaskId(numberOfTasks, numberOfWorkers, worker.ID, workerTaskId)

				logboek.Context(workerContext).Debug().LogF("Running worker %d with context %p for task %d/%d (%d)\n", worker.ID, workerContext, workerTaskId+1, numberOfTasksPerWorker[worker.ID], numberOfTasks)

				// Use channel to be able to cancel task immediately
				errCh := lo.Async(func() error {
					return taskFunc(workerContext, taskId)
				})

				// Block until one of next things is happened
				select {
				case <-groupCtx.Done():
					return context.Cause(groupCtx)
				// The task returns err or nil. When task returns err,
				// we mark worker as failed using failedWorkerCh and stop execution returning err.
				// The returned error is captured (via errgroup).
				case err = <-errCh:
					if err != nil {
						failedWorkerCh <- worker.ID
						return err
					}
				}
			}

			return nil
		})
	}

	g.Go(func() error {
		return printEachWorkerOutput(groupCtx, workers, failedWorkerCh)
	})

	return g.Wait()
}

func printEachWorkerOutput(ctx context.Context, workers []*parallelWorker, failedWorkerCh <-chan int) error {
	for _, worker := range workers {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case failedWorkerIdx := <-failedWorkerCh:
			return printWorkerOutput(ctx, workers[failedWorkerIdx])
		default:
			if err := printWorkerOutput(ctx, worker); err != nil {
				return err
			}
		}
	}

	return nil
}

func printWorkerOutput(ctx context.Context, worker *parallelWorker) error {
	var offset int64
	var err error

	for {
		select {
		case <-ctx.Done():
			return nil
		// final log flushing
		case <-worker.Writer.Done():
			_, err = copyOutput(ctx, worker.Reader.NewSectionReader(offset, math.MaxInt64))
			if err != nil {
				return fmt.Errorf("failed to copy worker final output: %w", err)
			}

			logboek.Context(ctx).LogOptionalLn()
			return nil
		// intermediate log flushing
		default:
			offset, err = copyOutput(ctx, worker.Reader.NewSectionReader(offset, 4096))
			if err != nil {
				return fmt.Errorf("failed to copy worker intermeediate output: %w", err)
			}

			if offset == 0 {
				time.Sleep(time.Millisecond * 100)
			}
		}
	}
}

func copyOutput(ctx context.Context, reader io.Reader) (int64, error) {
	var written int64
	var err error

	if err = logboek.Context(ctx).Streams().DoErrorWithoutIndent(func() error {
		written, err = io.Copy(logboek.Context(ctx).OutStream(), reader)
		return err
	}); err != nil {
		return 0, fmt.Errorf("failed to copy output: %w", err)
	}

	return written, nil
}

func calculateTaskId(tasksNumber, workersNumber, workerInd, workerTaskId int) int {
	taskId := workerInd*(tasksNumber/workersNumber) + workerTaskId

	rest := tasksNumber % workersNumber
	if rest != 0 {
		if rest > workerInd {
			taskId += workerInd
		} else {
			taskId += rest
		}
	}

	return taskId
}

func calculateTasksDistribution(numberOfTasks, maxNumberOfWorkers int) (int, []int) {
	numberOfWorkers := maxNumberOfWorkers
	if numberOfWorkers <= 0 || numberOfWorkers > numberOfTasks {
		numberOfWorkers = numberOfTasks
	}

	var numberOfTasksPerWorker []int
	for i := 0; i < numberOfWorkers; i++ {
		workerNumberOfTasks := numberOfTasks / numberOfWorkers
		rest := numberOfTasks % numberOfWorkers
		if rest > i {
			workerNumberOfTasks += 1
		}
		numberOfTasksPerWorker = append(numberOfTasksPerWorker, workerNumberOfTasks)
	}

	return numberOfWorkers, numberOfTasksPerWorker
}
