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

	numberOfWorkers, numberOfTasksPerWorker := calculateTasksDistribution(numberOfTasks, options.MaxNumberOfWorkers)

	return runWorkers(ctx, numberOfWorkers, options, func(workerCtx context.Context, worker *Worker) error {
		for workerTaskId := 0; workerTaskId < numberOfTasksPerWorker[worker.ID]; workerTaskId++ {
			select {
			case <-workerCtx.Done():
				logboek.Context(ctx).Debug().LogF("parallel: canceling worker %d with ctx %p for task %d/%d (%d)\n", worker.ID, workerCtx, workerTaskId, numberOfTasksPerWorker[worker.ID], numberOfTasks)
				return workerCtx.Err()
			default:
				taskId := calculateTaskId(numberOfTasks, numberOfWorkers, worker.ID, workerTaskId)
				logboek.Context(ctx).Debug().LogF("parallel: running worker %d with ctx %p for task %d/%d (%d)\n", worker.ID, workerCtx, workerTaskId, numberOfTasksPerWorker[worker.ID], numberOfTasks)

				if err := taskFunc(workerCtx, taskId); err != nil {
					return NewWorkerError(worker.ID, err)
				}
			}
		}

		return nil
	})
}

// NextTaskFunc returns the next task to run. It may block until a task
// becomes runnable. ok=false is terminal: the calling worker returns and is
// never asked again.
type NextTaskFunc func(ctx context.Context) (taskId int, ok bool, err error)

// DoTasksDynamic runs workers that each repeatedly pull the next task to run
// from `next` (instead of a fixed, statically-partitioned task range like
// DoTasks) until `next` reports there's nothing left. This allows the caller
// to drive a dynamic dependency-graph scheduler where the set of runnable
// tasks isn't known upfront and grows as earlier tasks complete.
//
// options.MaxNumberOfWorkers <= 0 means a single worker (the task count is
// unknown upfront), unlike DoTasks where it means one worker per task.
func DoTasksDynamic(ctx context.Context, options DoTasksOptions, next NextTaskFunc, taskFunc TaskFunc) error {
	numberOfWorkers := options.MaxNumberOfWorkers
	if numberOfWorkers <= 0 {
		numberOfWorkers = 1
	}

	logboek.Context(ctx).Debug().LogF("parallel: initializing dynamic scheduler with %d workers\n", numberOfWorkers)

	return runWorkers(ctx, numberOfWorkers, options, func(workerCtx context.Context, worker *Worker) error {
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

func runWorkers(ctx context.Context, numberOfWorkers int, options DoTasksOptions, workerLoop func(workerCtx context.Context, worker *Worker) error) error {
	groupParentCtx, cancelGroupParentCtx := context.WithCancel(ctx)
	defer cancelGroupParentCtx()

	g, groupCtx := errgroup.WithContext(groupParentCtx)

	workers := make([]*Worker, 0, numberOfWorkers)
	workerCtxs := make([]context.Context, 0, numberOfWorkers)

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

	// All workers and their contexts are created before any goroutine starts,
	// so an initialization failure returns with nothing spawned.
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

		workerCtxs = append(workerCtxs, workerCtx)
	}

	for i, worker := range workers {
		workerCtx := workerCtxs[i]

		g.Go(func() error {
			defer func() {
				if err := worker.HalfClose(); err != nil {
					logboek.Context(ctx).Warn().LogF("parallel: failed to half-close worker %d: %s\n", worker.ID, err)
				}
			}()

			return workerLoop(workerCtx, worker)
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
