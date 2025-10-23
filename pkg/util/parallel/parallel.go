package parallel

import (
	"context"
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

	// Create slice of buffered workers
	workers := make([]*bufferedWorker, numberOfWorkers)
	// Create channel to signal failed worker
	failedWorkerCh := make(chan int, 1)

	for i := 0; i < numberOfWorkers; i++ {
		worker := newBufferedWorker(i)
		workers[worker.ID] = worker

		// Create a new context with a background task ID for each worker
		ctxWithBackgroundTaskID := context.WithValue(groupCtx, CtxBackgroundTaskIDKey, worker.ID)
		workerContext := logboek.NewContext(ctxWithBackgroundTaskID, logging.NewSubLogger(ctxWithBackgroundTaskID, worker.Buffer(), worker.Buffer()))

		if options.InitDockerCLIForEachWorker {
			var err error
			// TODO: should we always create new docker context for each worker to prevent "context canceled" error?
			if workerContext, err = docker.NewContext(workerContext); err != nil {
				return err
			}
		}

		g.Go(func() error {
			defer worker.MarkDone()

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
				case err := <-errCh:
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
		return printEachWorkerOutput(ctx, workers, failedWorkerCh)
	})

	return g.Wait()
}

// printEachWorkerOutput prints the output of each buffered worker in the provided slice,
// while monitoring for context cancellation and failed workers.
func printEachWorkerOutput(ctx context.Context, bufferedWorkers []*bufferedWorker, failedWorkerCh <-chan int) error {
	for i := 0; i < len(bufferedWorkers); i++ {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case failedWorkerIdx := <-failedWorkerCh:
			return printWorkerOutput(ctx, bufferedWorkers[failedWorkerIdx])
		default:
			if err := printWorkerOutput(ctx, bufferedWorkers[i]); err != nil {
				return err
			}
		}
	}

	return nil
}

// printWorkerOutput prints the output of a buffered worker to the provided context's output stream.
// It continuously reads from the worker's buffer and writes it to the output stream until the worker is done.
// If an error occurs during the copy process, it is returned immediately.
func printWorkerOutput(ctx context.Context, worker *bufferedWorker) error {
	for {
		var n int64
		var err error
		if err = logboek.Context(ctx).Streams().DoErrorWithoutIndent(func() error {
			n, err = io.Copy(logboek.Context(ctx).OutStream(), worker.Buffer())
			return err
		}); err != nil {
			return err
		}

		if worker.IsDone() {
			logboek.Context(ctx).LogOptionalLn()
			return nil
		}

		if n == 0 {
			time.Sleep(time.Millisecond * 100)
		}
	}
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
