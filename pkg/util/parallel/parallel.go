package parallel

import (
	"context"
	"fmt"
	"io"
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

	workers := make([]*Worker, 0, numberOfWorkers)

	defer func() {
		for _, worker := range workers {
			if err := worker.Close(); err != nil {
				logboek.Context(ctx).Warn().LogF("Failed to close worker %d: %s\n", worker.ID, err)
			}
			if err := worker.Cleanup(); err != nil {
				logboek.Context(ctx).Warn().LogF("Failed to cleanup worker %d: %s\n", worker.ID, err)
			}
		}
	}()

	for i := 0; i < numberOfWorkers; i++ {
		worker, err := NewWorker(i)
		if err != nil {
			return fmt.Errorf("failed to create worker %d: %w", i, err)
		}
		workers = append(workers, worker)

		// Create a new context with a background task ID for each worker
		taskIDCtx := context.WithValue(groupCtx, CtxBackgroundTaskIDKey, worker.ID)
		workerCtx := logboek.NewContext(taskIDCtx, logging.NewSubLogger(taskIDCtx, worker, worker))

		if options.InitDockerCLIForEachWorker {
			// TODO: should we always create new docker context for each worker to prevent "context canceled" error?
			if workerCtx, err = docker.NewContext(workerCtx); err != nil {
				return err
			}
		}

		g.Go(func() error {
			defer worker.HalfClose()

			for workerTaskId := 0; workerTaskId < numberOfTasksPerWorker[worker.ID]; workerTaskId++ {
				taskId := calculateTaskId(numberOfTasks, numberOfWorkers, worker.ID, workerTaskId)

				logboek.Context(workerCtx).Debug().LogF("Running worker %d with context %p for task %d/%d (%d)\n", worker.ID, workerCtx, workerTaskId+1, numberOfTasksPerWorker[worker.ID], numberOfTasks)

				// Use channel to be able to cancel task immediately
				errCh := lo.Async(func() error {
					return taskFunc(workerCtx, taskId)
				})

				select {
				case <-workerCtx.Done():
					return context.Cause(workerCtx)
				// The taskFunc returns err or nil. On err we mark worker as failed.
				// The returned error is captured (via errgroup).
				case err = <-errCh:
					if err != nil {
						worker.Fail()
						return err
					}
				}
			}

			return nil
		})
	}

	g.Go(func() error {
		return printEachWorkerOutput(groupCtx, workers)
	})

	return g.Wait()
}

func printEachWorkerOutput(ctx context.Context, workers []*Worker) error {
	for _, worker := range workers {
		select {
		case <-ctx.Done():
			// If failed worker is found, print its output
			for _, w := range workers {
				if w.Failed() {
					return printWorkerOutput(context.WithoutCancel(ctx), w)
				}
			}
			// Otherwise, return the error from ctx
			return context.Cause(ctx)
		default:
			if err := printWorkerOutput(ctx, worker); err != nil {
				return err
			}
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
			return nil
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

	return nil
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
