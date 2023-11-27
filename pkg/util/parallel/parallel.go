package parallel

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/util/parallel/constant"
)

type DoTasksOptions struct {
	InitDockerCLIForEachWorker bool
	MaxNumberOfWorkers         int
	LiveOutput                 bool
}

func DoTasks(ctx context.Context, numberOfTasks int, options DoTasksOptions, taskFunc func(ctx context.Context, taskId int) error) error {
	if numberOfTasks == 0 {
		return nil
	}

	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	// determine number of tasks
	numberOfWorkers := options.MaxNumberOfWorkers
	if numberOfWorkers <= 0 || numberOfWorkers > numberOfTasks {
		numberOfWorkers = numberOfTasks
	}

	// distribute tasks among workers
	var numberOfTasksPerWorker []int
	for i := 0; i < numberOfWorkers; i++ {
		workerNumberOfTasks := numberOfTasks / numberOfWorkers
		rest := numberOfTasks % numberOfWorkers
		if rest > i {
			workerNumberOfTasks += 1
		}

		numberOfTasksPerWorker = append(numberOfTasksPerWorker, workerNumberOfTasks)
	}

	taskResultFailedCh := make(chan *bufWorkerTaskResult)
	taskResultDoneCh := make(chan *bufWorkerTaskResult)
	workerDoneCh := make(chan *bufWorker)
	quitCh := make(chan bool)

	var workers []*bufWorker
	for i := 0; i < numberOfWorkers; i++ {
		var workerContext context.Context

		workerID := i
		workerBuf := &util.GoroutineSafeBuffer{Buffer: bytes.NewBuffer([]byte{})}
		worker := &bufWorker{buf: workerBuf}
		workers = append(workers, worker)

		ctxWithBackgroundTaskID := context.WithValue(ctx, constant.CtxBackgroundTaskIDKey, workerID)
		workerContext = logboek.NewContext(ctxWithBackgroundTaskID, logboek.Context(ctx).NewSubLogger(workerBuf, workerBuf))
		{
			logboek.Context(workerContext).Streams().SetPrefixStyle(style.Highlight())
			if logboek.Context(workerContext).Streams().IsPrefixTimeEnabled() {
				logboek.Context(workerContext).Streams().SetPrefixTimeFormat("15:04:05")
			}

			if options.InitDockerCLIForEachWorker {
				workerContextWithDockerCli, err := docker.NewContext(workerContext)
				if err != nil {
					return err
				}

				workerContext = workerContextWithDockerCli
			}
		}

		go func() {
			workerNumberOfTasks := numberOfTasksPerWorker[workerID]
			for workerTaskId := 0; workerTaskId < workerNumberOfTasks; workerTaskId++ {
				taskId := calculateTaskId(numberOfTasks, numberOfWorkers, workerID, workerTaskId)
				if debug() {
					logboek.Context(workerContext).LogF("Running worker %d task %d/%d (%d)\n", workerID, workerTaskId+1, workerNumberOfTasks, numberOfTasks)
				}
				err := taskFunc(workerContext, taskId)

				ch := taskResultDoneCh
				if err != nil {
					ch = taskResultFailedCh
				}

				select {
				case ch <- worker.TaskResult(err):
					if err != nil {
						return
					}
				case <-quitCh:
					return
				}
			}

			workerDoneCh <- worker
		}()
	}

	var err error
	if options.LiveOutput {
		err = workersHandlerLiveOutput(ctx, workers, taskResultDoneCh, taskResultFailedCh, quitCh, workerDoneCh)
	} else {
		err = workersHandlerStandard(ctx, workers, taskResultDoneCh, taskResultFailedCh, quitCh, workerDoneCh)
	}

	return err
}

func workersHandlerLiveOutput(ctx context.Context, workers []*bufWorker, taskResultDoneCh, taskResultFailedCh chan *bufWorkerTaskResult, quitCh chan bool, workerDoneCh chan *bufWorker) error {
workerLoop:
	for _, currentWorker := range workers {
		for {
			select {
			case <-taskResultDoneCh:
			case taskResult := <-taskResultFailedCh:
				close(quitCh)

				if taskResult.worker != currentWorker {
					logboek.Context(ctx).LogLn()
				}

				if err := logboek.Context(ctx).Streams().DoErrorWithoutIndent(func() error {
					_, err := io.Copy(logboek.Context(ctx).OutStream(), taskResult.worker.buf)
					return err
				}); err != nil {
					return err
				}

				logboek.Context(ctx).LogOptionalLn()

				return taskResult.err
			case worker := <-workerDoneCh:
				worker.isDone = true
			default:
				var n int64
				var err error
				if err := logboek.Context(ctx).Streams().DoErrorWithoutIndent(func() error {
					n, err = io.Copy(logboek.Context(ctx).OutStream(), currentWorker.buf)
					return err
				}); err != nil {
					return err
				}

				if currentWorker.isDone {
					logboek.Context(ctx).LogOptionalLn()
					continue workerLoop
				}

				if n == 0 {
					time.Sleep(time.Millisecond * 100)
				}
			}
		}
	}

	return nil
}

func workersHandlerStandard(ctx context.Context, workers []*bufWorker, taskResultDoneCh, taskResultFailedCh chan *bufWorkerTaskResult, quitCh chan bool, workerDoneCh chan *bufWorker) error {
	var workerDoneCounter int
	for {
		select {
		case taskResult := <-taskResultDoneCh:
			if err := logboek.Context(ctx).Streams().DoErrorWithoutIndent(func() error {
				_, err := io.Copy(logboek.Context(ctx).OutStream(), taskResult.worker.buf)
				return err
			}); err != nil {
				return err
			}

			logboek.Context(ctx).LogOptionalLn()
		case taskResult := <-taskResultFailedCh:
			close(quitCh)

			if err := logboek.Context(ctx).Streams().DoErrorWithoutIndent(func() error {
				_, err := io.Copy(logboek.Context(ctx).OutStream(), taskResult.worker.buf)
				return err
			}); err != nil {
				return err
			}

			logboek.Context(ctx).LogOptionalLn()

			return taskResult.err
		case <-workerDoneCh:
			workerDoneCounter++
			if workerDoneCounter == len(workers) {
				return nil
			}
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

func debug() bool {
	return os.Getenv("WERF_DEBUG_PARALLEL") == "1"
}
