package parallel

import (
	"bytes"
	"context"
	"os"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/util/parallel/constant"
)

type DoTasksOptions struct {
	InitDockerCLIForEachWorker bool
	MaxNumberOfWorkers         int
	IsLiveOutputOn             bool
}

func DoTasks(ctx context.Context, numberOfTasks int, options DoTasksOptions, taskFunc func(ctx context.Context, taskId int) error) error {
	if numberOfTasks == 0 {
		return nil
	}

	numberOfWorkers := options.MaxNumberOfWorkers
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

	errCh := make(chan interface{})
	doneTaskCh := make(chan interface{})
	doneWorkerCh := make(chan worker)
	quitCh := make(chan bool)
	doneWorkersCounter := numberOfWorkers
	isLiveOutputOnFlag := options.IsLiveOutputOn

	var liveLogger types.LoggerInterface
	var liveContext context.Context
	if options.IsLiveOutputOn {
		liveLogger = logboek.NewLogger(os.Stdout, os.Stderr)
		liveLogger.GetStreamsSettingsFrom(logboek.Context(ctx))
		liveLogger.SetAcceptedLevel(logboek.Context(ctx).AcceptedLevel())

		liveContext = logboek.NewContext(ctx, liveLogger)

		if docker.IsContext(liveContext) {
			if err := docker.SyncContextCliWithLogger(liveContext); err != nil {
				return err
			}
			defer docker.SyncContextCliWithLogger(ctx)
		}
	}

	var workersBuffs []*bytes.Buffer
	var doneTaskDataList [][]byte
	for i := 0; i < numberOfWorkers; i++ {
		var workerContext context.Context
		var worker worker

		workerId := i

		if i == 0 && options.IsLiveOutputOn {
			workerContext = liveContext
			worker = &liveWorker{}
		} else {
			workerBuf := bytes.NewBuffer([]byte{})
			workersBuffs = append(workersBuffs, workerBuf)
			worker = &bufWorker{buf: workerBuf}

			ctxWithBackgroundTaskID := context.WithValue(ctx, constant.CtxBackgroundTaskIDKey, i)
			workerContext = logboek.NewContext(ctxWithBackgroundTaskID, logboek.Context(ctx).NewSubLogger(workerBuf, workerBuf))
			logboek.Context(workerContext).Streams().SetPrefixStyle(style.Highlight())

			if options.InitDockerCLIForEachWorker {
				workerContextWithDockerCli, err := docker.NewContext(workerContext)
				if err != nil {
					return err
				}

				workerContext = workerContextWithDockerCli
			}
		}

		go func() {
			workerNumberOfTasks := numberOfTasksPerWorker[workerId]

			for workerTaskId := 0; workerTaskId < workerNumberOfTasks; workerTaskId++ {
				taskId := calculateTaskId(numberOfTasks, numberOfWorkers, workerId, workerTaskId)
				if debug() {
					logboek.Context(workerContext).LogF("Running worker %d task %d/%d (%d)\n", workerId, workerTaskId+1, workerNumberOfTasks, numberOfTasks)
				}
				err := taskFunc(workerContext, taskId)

				ch := doneTaskCh
				if err != nil {
					ch = errCh
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

			doneWorkerCh <- worker
		}()
	}

	for {
		select {
		case res := <-doneTaskCh:
			switch taskResult := res.(type) {
			case *bufWorkerTaskResult:
				if isLiveOutputOnFlag {
					doneTaskDataList = append(doneTaskDataList, taskResult.data)
				} else {
					processTaskResultData(ctx, taskResult.data)
				}
			}
		case res := <-errCh:
			close(quitCh)

			switch taskResult := res.(type) {
			case *bufWorkerTaskResult:
				if isLiveOutputOnFlag {
					liveLogger.Streams().Mute()
				}

				for _, data := range doneTaskDataList {
					processTaskResultData(ctx, data)
				}

				for _, buf := range workersBuffs {
					if buf != taskResult.buf {
						processTaskResultData(ctx, []byte(buf.String()))
					}
				}

				processTaskResultData(ctx, taskResult.data)

				return taskResult.err
			case *lifeWorkerTaskResult:
				if len(workersBuffs) != 0 {
					if logboek.Context(ctx).Info().IsAccepted() {
						logboek.Context(liveContext).LogLn()

						for _, data := range doneTaskDataList {
							processTaskResultData(ctx, data)
						}

						for _, buf := range workersBuffs {
							processTaskResultData(ctx, buf.Bytes())
						}
					}
				}

				return taskResult.err
			}
		case res := <-doneWorkerCh:
			if res.IsLiveWorker() {
				isLiveOutputOnFlag = false

				for _, data := range doneTaskDataList {
					processTaskResultData(ctx, data)
				}
			}

			doneWorkersCounter--
			if doneWorkersCounter == 0 {
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

func processTaskResultData(ctx context.Context, data []byte) {
	if len(data) == 0 { // TODO: fix in logboek
		return
	}

	logboek.Streams().DoWithoutIndent(func() {
		_, _ = logboek.Context(ctx).OutStream().Write(data)
		logboek.Context(ctx).LogOptionalLn()
	})
}

func debug() bool {
	return os.Getenv("WERF_DEBUG_PARALLEL") == "1"
}
