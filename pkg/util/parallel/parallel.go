package parallel

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"

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

	return runWorkers(ctx, numberOfWorkers, options, taskFunc, func(workerCtx context.Context, worker *Worker, runTask func(taskId int) error) error {
		for workerTaskId := 0; workerTaskId < numberOfTasksPerWorker[worker.ID]; workerTaskId++ {
			select {
			case <-workerCtx.Done():
				logboek.Context(ctx).Debug().LogF("parallel: canceling worker %d with ctx %p for task %d/%d (%d)\n", worker.ID, workerCtx, workerTaskId, numberOfTasksPerWorker[worker.ID], numberOfTasks)
				return workerCtx.Err()
			default:
				taskId := calculateTaskId(numberOfTasks, numberOfWorkers, worker.ID, workerTaskId)
				logboek.Context(ctx).Debug().LogF("parallel: running worker %d with ctx %p for task %d/%d (%d)\n", worker.ID, workerCtx, workerTaskId, numberOfTasksPerWorker[worker.ID], numberOfTasks)

				if err := runTask(taskId); err != nil {
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

	return runWorkers(ctx, numberOfWorkers, options, taskFunc, func(workerCtx context.Context, worker *Worker, runTask func(taskId int) error) error {
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

			if err := runTask(taskId); err != nil {
				return NewWorkerError(worker.ID, err)
			}
		}
	})
}

// runWorkers hands each workerLoop a runTask that binds the task's logger to
// its own TaskOutput and registers it with the Printer, so the loops only
// decide WHICH task to run next. The worker context carries the worker ID
// and, when requested, a docker cli whose output is relayed to the logger
// of the worker's current task (see Worker).
func runWorkers(ctx context.Context, numberOfWorkers int, options DoTasksOptions, taskFunc TaskFunc, workerLoop func(workerCtx context.Context, worker *Worker, runTask func(taskId int) error) error) error {
	groupParentCtx, cancelGroupParentCtx := context.WithCancel(ctx)
	defer cancelGroupParentCtx()

	g, groupCtx := errgroup.WithContext(groupParentCtx)

	printer := NewPrinter()
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
	// so an initialization failure returns with nothing spawned. The worker
	// context carries a template logger that nothing ever writes to: task
	// loggers are cloned from it, because cloning reads the parent's stream
	// state and the caller's logger is being written to by the printer for
	// the whole run.
	for i := 0; i < numberOfWorkers; i++ {
		worker := NewWorker(i)
		workers = append(workers, worker)

		workerCtx := context.WithValue(groupCtx, CtxBackgroundTaskIDKey, worker.ID)
		workerCtx = logboek.NewContext(workerCtx, logging.NewSubLogger(workerCtx, io.Discard, io.Discard))

		if options.InitDockerCLIForEachWorker {
			var err error
			if workerCtx, err = docker.NewContextWithStreams(workerCtx, worker.OutStream(), worker.ErrStream()); err != nil {
				return err
			}
		}

		workerCtxs = append(workerCtxs, workerCtx)
	}

	var runningWorkers atomic.Int32
	runningWorkers.Store(int32(numberOfWorkers))
	if numberOfWorkers == 0 {
		printer.Close()
	}

	// A worker may only look for its first task once the previous worker has
	// started (or given up on) its own, so the first task of each worker is
	// enqueued in worker order rather than in goroutine-scheduling order.
	// Everything after a worker's first task is ordered by real start time.
	firstTaskStarted := make([]chan struct{}, numberOfWorkers)
	for i := range firstTaskStarted {
		firstTaskStarted[i] = make(chan struct{})
	}

	for i, worker := range workers {
		workerCtx := workerCtxs[i]

		g.Go(func() error {
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(firstTaskStarted[i]) }) }

			defer func() {
				release()
				if err := worker.endTask(); err != nil {
					logboek.Context(ctx).Warn().LogF("parallel: failed to half-close worker %d output: %s\n", worker.ID, err)
				}
				if runningWorkers.Add(-1) == 0 {
					printer.Close()
				}
			}()

			if i > 0 {
				select {
				case <-firstTaskStarted[i-1]:
				case <-workerCtx.Done():
					return workerCtx.Err()
				}
			}

			return workerLoop(workerCtx, worker, func(taskId int) error {
				out := worker.beginTask()

				startOrder := printer.Enqueue(out)
				release()

				taskCtx := context.WithValue(workerCtx, CtxTaskStartOrderKey, startOrder)
				taskLogger := logging.NewSubLogger(taskCtx, out, out)
				// Cloning subtracts the indentation again; the template already
				// paid it once, so the task logger gets the template's width back.
				taskLogger.Streams().SetWidth(logboek.Context(workerCtx).Streams().Width())
				taskCtx = logboek.NewContext(taskCtx, taskLogger)
				worker.bindTaskStreams(taskLogger.OutStream(), taskLogger.ErrStream())

				defer func() {
					// logboek holds an incomplete line back until the next write;
					// LogF("") is that write, so the line reaches the buffer before
					// HalfClose terminates it.
					taskLogger.LogF("")
					if err := worker.endTask(); err != nil {
						logboek.Context(ctx).Warn().LogF("parallel: failed to half-close task %d output: %s\n", taskId, err)
					}
				}()

				if err := taskFunc(taskCtx, taskId); err != nil {
					worker.failTask(out)
					return err
				}

				return nil
			})
		})
	}

	g.Go(func() error {
		return printer.Print(groupCtx)
	})

	if err := g.Wait(); err != nil {
		// There are two cases how to continue printing:
		// 1. Receiving the system signal (SIGINT / SIGTERM). We detect it by checking "context canceled" error.
		// 	- We continue to print the queue as is.
		// 2. Getting an error from a task. We detect it by checking non "context canceled" error.
		//	- The failed task is moved to the end of the printing queue (to highlight the error to the user),
		//	  unless it is the one being printed right now — then the tasks queued behind it are discarded.
		//    An error raised outside a task leaves the queue as is: no block is to blame for it.

		if !isCanceledErr(err) {
			var workerErr *WorkerError

			if errors.As(err, &workerErr) {
				printer.FailFast(workers[workerErr.ID].failedOutput())
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
