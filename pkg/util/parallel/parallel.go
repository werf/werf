package parallel

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	logboek.Context(ctx).Debug().LogF("parallel: initializing with options %d (workers) per %d (tasks)\n", options.MaxNumberOfWorkers, numberOfTasks)

	g, groupCtx := errgroup.WithContext(ctx)

	// Determine number of workers and tasks per worker
	numberOfWorkers, numberOfTasksPerWorker := calculateTasksDistribution(numberOfTasks, options.MaxNumberOfWorkers)

	workers := make([]*Worker, 0, numberOfWorkers)

	defer func() {
		for _, worker := range workers {
			if err := worker.Close(); err != nil {
				logboek.Context(ctx).Warn().LogF("parallel: failed to close worker %d: %s\n", worker.ID, err)
			}
			if err := worker.Cleanup(); err != nil {
				logboek.Context(ctx).Warn().LogF("parallel: failed to cleanup worker %d: %s\n", worker.ID, err)
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
			if workerCtx, err = docker.NewContext(workerCtx); err != nil {
				return err
			}
		}

		g.Go(func() error {
			defer func() {
				if err = worker.HalfClose(); err != nil {
					logboek.Context(ctx).Warn().LogF("parallel: failed to half-close worker %d: %s\n", worker.ID, err)
				}
			}()

			for workerTaskId := 0; workerTaskId < numberOfTasksPerWorker[worker.ID]; workerTaskId++ {
				select {
				case <-workerCtx.Done():
					logboek.Context(ctx).Debug().LogF("parallel: canceling worker %d with ctx %p for task %d/%d (%d)\n", worker.ID, workerCtx, workerTaskId, numberOfTasksPerWorker[worker.ID], numberOfTasks)
					return workerCtx.Err()
				default:
					taskId := calculateTaskId(numberOfTasks, numberOfWorkers, worker.ID, workerTaskId)
					logboek.Context(ctx).Debug().LogF("parallel: running worker %d with ctx %p for task %d/%d (%d)\n", worker.ID, workerCtx, workerTaskId, numberOfTasksPerWorker[worker.ID], numberOfTasks)

					if err = taskFunc(workerCtx, taskId); err != nil {
						return NewWorkerError(worker.ID, err)
					}
				}
			}

			return nil
		})
	}

	printer := NewPrinter(workers)

	g.Go(func() error {
		return printer.Print(groupCtx)
	})

	if err := g.Wait(); err != nil {
		// There are two cases how to continue printing:
		// 1. Receiving the system signal (SIGINT / SIGTERM). We detect it by checking "context canceled" error.
		// 	- We continue to print starting from 'foreground' worker through the rest workers without any changes.
		// 2. Getting an error from a worker. We detect it by checking non "context canceled" error.
		//	- If 'foreground' worker IS NOT THE SAME worker which returned the error,
		//	  we move errored worker to the end of the printing queue (to highlight the error to the user)
		//    and we continue to print starting from 'foreground' through the rest workers.
		//  - If 'foreground' worker IS THE SAME worker which returned the error,
		// 	  we continue to print starting from 'foreground' (errored) worker,
		//    and we discard logs from the rest workers.

		if !isCanceledErr(err) {
			var workerErr *WorkerError

			if errors.As(err, &workerErr) {
				if printer.Cur() != workerErr.ID {
					printer.Swap(printer.Max(), workerErr.ID) // move filed worker to the end of the printing queue
				} else {
					printer.SetMax(printer.Cur()) // discard logs from the rest workers
				}
			}
		}

		err1 := printer.Print(context.WithoutCancel(ctx))

		return errors.Join(err, err1)
	}

	return nil
}

// NextTaskFunc returns the next task to run. ok is false when there are no
// more tasks left to schedule right now (the caller may be asked again later,
// once in-flight tasks free up new work, until ctx is done).
type NextTaskFunc func(ctx context.Context) (taskId int, ok bool, err error)

// DoTasksDynamic runs numberOfWorkers workers that each repeatedly pull the
// next task to run from `next` (instead of a fixed, statically-partitioned
// task range like DoTasks) until `next` reports there's nothing left. This
// allows the caller to drive a dynamic dependency-graph scheduler where the
// set of runnable tasks isn't known upfront and grows as earlier tasks
// complete.
func DoTasksDynamic(ctx context.Context, numberOfWorkers int, options DoTasksOptions, next NextTaskFunc, taskFunc TaskFunc) error {
	if numberOfWorkers <= 0 {
		numberOfWorkers = 1
	}

	logboek.Context(ctx).Debug().LogF("parallel: initializing dynamic scheduler with %d workers\n", numberOfWorkers)

	g, groupCtx := errgroup.WithContext(ctx)

	workers := make([]*Worker, 0, numberOfWorkers)

	defer func() {
		for _, worker := range workers {
			if err := worker.Close(); err != nil {
				logboek.Context(ctx).Warn().LogF("parallel: failed to close worker %d: %s\n", worker.ID, err)
			}
			if err := worker.Cleanup(); err != nil {
				logboek.Context(ctx).Warn().LogF("parallel: failed to cleanup worker %d: %s\n", worker.ID, err)
			}
		}
	}()

	for i := 0; i < numberOfWorkers; i++ {
		worker, err := NewWorker(i)
		if err != nil {
			return fmt.Errorf("failed to create worker %d: %w", i, err)
		}
		workers = append(workers, worker)

		taskIDCtx := context.WithValue(groupCtx, CtxBackgroundTaskIDKey, worker.ID)
		workerCtx := logboek.NewContext(taskIDCtx, logging.NewSubLogger(taskIDCtx, worker, worker))

		if options.InitDockerCLIForEachWorker {
			if workerCtx, err = docker.NewContext(workerCtx); err != nil {
				return err
			}
		}

		g.Go(func() error {
			defer func() {
				if err = worker.HalfClose(); err != nil {
					logboek.Context(ctx).Warn().LogF("parallel: failed to half-close worker %d: %s\n", worker.ID, err)
				}
			}()

			for {
				select {
				case <-workerCtx.Done():
					logboek.Context(ctx).Debug().LogF("parallel: canceling worker %d with ctx %p\n", worker.ID, workerCtx)
					return workerCtx.Err()
				default:
				}

				taskId, ok, err := next(workerCtx)
				if err != nil {
					return NewWorkerError(worker.ID, err)
				}
				if !ok {
					return nil
				}

				logboek.Context(ctx).Debug().LogF("parallel: running worker %d with ctx %p for task %d\n", worker.ID, workerCtx, taskId)

				if err := taskFunc(workerCtx, taskId); err != nil {
					return NewWorkerError(worker.ID, err)
				}
			}
		})
	}

	printer := NewPrinter(workers)

	g.Go(func() error {
		return printer.Print(groupCtx)
	})

	if err := g.Wait(); err != nil {
		if !isCanceledErr(err) {
			var workerErr *WorkerError

			if errors.As(err, &workerErr) {
				if printer.Cur() != workerErr.ID {
					printer.Swap(printer.Max(), workerErr.ID)
				} else {
					printer.SetMax(printer.Cur())
				}
			}
		}

		err1 := printer.Print(context.WithoutCancel(ctx))

		return errors.Join(err, err1)
	}

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

// isCanceledErr is a workaround to check "context canceled" error from docker daemon
func isCanceledErr(err error) bool {
	return strings.HasSuffix(err.Error(), context.Canceled.Error())
}
