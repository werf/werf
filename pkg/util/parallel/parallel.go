package parallel

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/werf/v2/pkg/docker"
)

type DoTasksOptions struct {
	InitDockerCLIForEachWorker bool
	MaxNumberOfWorkers         int
}

func DoTasks(ctx context.Context, numberOfTasks int, options DoTasksOptions, taskFunc func(ctx context.Context, taskId int) error) error {
	if numberOfTasks == 0 {
		return nil
	}

	// Use errgroup to sync workerBuffers and propagate errors
	g, groupCtx := errgroup.WithContext(ctx)

	// Determine number of workers and tasks per worker
	numberOfWorkers, numberOfTasksPerWorker := calculateTasksDistribution(numberOfTasks, options.MaxNumberOfWorkers)

	// Create buffers for each worker to log output
	workerBuffers := make([]*util.GoroutineSafeBuffer, numberOfWorkers)

	for i := 0; i < numberOfWorkers; i++ {
		workerID := i
		workerBuffers[workerID] = &util.GoroutineSafeBuffer{Buffer: bytes.NewBuffer([]byte{})}

		// Create a new context with a background task ID for each worker
		ctxWithBackgroundTaskID := context.WithValue(groupCtx, CtxBackgroundTaskIDKey, workerID)
		workerContext := logboek.NewContext(ctxWithBackgroundTaskID, logboek.Context(ctxWithBackgroundTaskID).NewSubLogger(workerBuffers[workerID], workerBuffers[workerID]))

		logboek.Context(workerContext).Streams().SetPrefixStyle(style.Highlight())
		if logboek.Context(workerContext).Streams().IsPrefixTimeEnabled() {
			logboek.Context(workerContext).Streams().SetPrefixTimeFormat("15:04:05")
		}

		if options.InitDockerCLIForEachWorker {
			var err error
			// TODO: should we always create new docker context for each worker to prevent "context canceled" error?
			if workerContext, err = docker.NewContext(workerContext); err != nil {
				return err
			}
		}

		g.Go(func() error {
			for workerTaskId := 0; workerTaskId < numberOfTasksPerWorker[workerID]; workerTaskId++ {
				taskId := calculateTaskId(numberOfTasks, numberOfWorkers, workerID, workerTaskId)
				if debug() {
					logboek.Context(workerContext).LogF("Running worker %d task %d/%d (%d)\n", workerID, workerTaskId+1, numberOfTasksPerWorker[workerID], numberOfTasks)
				}
				if err := taskFunc(workerContext, taskId); err != nil {
					return err
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return handleWorkerOutput(ctx, workerBuffers)
}

func handleWorkerOutput(ctx context.Context, workerBuffers []*util.GoroutineSafeBuffer) error {
	for _, buffer := range workerBuffers {
		for {
			var n int64
			var err error
			if err = logboek.Context(ctx).Streams().DoErrorWithoutIndent(func() error {
				n, err = io.Copy(logboek.Context(ctx).OutStream(), buffer)
				return err
			}); err != nil {
				return err
			}
			if n == 0 {
				break
			}
			time.Sleep(time.Millisecond * 100)
		}

		// Finished with worker output
		logboek.Context(ctx).LogOptionalLn()
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

func debug() bool {
	return os.Getenv("WERF_DEBUG_PARALLEL") == "1"
}
