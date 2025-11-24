package parallel_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/util/parallel"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = DescribeTable("parallel task",
	func(
		ctx context.Context,
		parallelExecutionLimit time.Duration,
		numberOfTasks int,
		options parallel.DoTasksOptions,
		spyTask *spyTaskFunc,
		expectedCallsCount int,
		expectedErrMatcher types.GomegaMatcher,
		expectedOutputLines []string,
	) {
		output := newSpyOutput(len(expectedOutputLines))
		ctx = logboek.NewContext(ctx, logboek.NewLogger(output, output))

		if parallelExecutionLimit > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, parallelExecutionLimit)
			defer cancel()
		}

		// tmp_manager requires werf init
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())

		err := parallel.DoTasks(ctx, numberOfTasks, options, spyTask.Callback)

		Expect(spyTask.Count()).To(Equal(expectedCallsCount))
		Expect(err).To(expectedErrMatcher)
		Expect(output.Lines()).To(HaveExactElements(expectedOutputLines))
	},
	Entry(
		"should do nothing if num_of_tasks=0",
		time.Duration(0),
		0,
		parallel.DoTasksOptions{},
		newSpyTask(func(ctx context.Context, taskId int) error {
			return errors.New("should not be called")
		}),
		0,
		Succeed(),
		[]string{},
	),
	Entry(
		"should print log concurrently while long task execution",
		time.Duration(0),
		4,
		parallel.DoTasksOptions{
			MaxNumberOfWorkers: 2,
		},
		newSpyTask(func(ctx context.Context, taskId int) error {
			switch taskId {
			case 0:
				time.Sleep(100 * time.Millisecond)
				logboek.Context(ctx).LogLn("one")
			case 1:
				time.Sleep(200 * time.Millisecond)
				logboek.Context(ctx).LogLn("two")
			case 2:
				time.Sleep(300 * time.Millisecond)
				logboek.Context(ctx).LogLn("three")
				logboek.Context(ctx).LogLn("four")
			case 3:
				time.Sleep(400 * time.Millisecond)
				logboek.Context(ctx).LogLn("five")
			}
			return nil
		}),
		4,
		Succeed(),
		[]string{
			"one\n",
			"two\n",
			"\nthree\nfour\n",
			"five\n",
		},
	),
	Entry(
		"should handle error from one of workers (fail fast) and stop execution via context cancellation for another workers",
		time.Duration(0),
		2,
		parallel.DoTasksOptions{
			MaxNumberOfWorkers: 2,
		},
		newSpyTask(func(ctx context.Context, taskId int) error {
			switch taskId {
			case 0:
				logboek.Context(ctx).LogLn("workers[0], task[0]: workers is active and it prints its log")
				<-ctx.Done()
				Expect(ctx.Err()).To(MatchError(context.Canceled))
				logboek.Context(ctx).LogLn("workers[0], task[0]: worker is still active and it finishes printing its log")
				return nil
			case 1:
				time.Sleep(150 * time.Millisecond)
				logboek.Context(ctx).LogLn("workers[1]: task[1]: worker was non-active and it was failed, because of that workers' log will be printed in the end")
				return errors.New("task 1 failed")
			default:
				panic(fmt.Sprintf("unexpected taskId: %d", taskId))
			}
		}),
		2,
		MatchError("task 1 failed"),
		[]string{
			"workers[0], task[0]: workers is active and it prints its log\n",
			"workers[0], task[0]: worker is still active and it finishes printing its log\n",
			"\nworkers[1]: task[1]: worker was non-active and it was failed, because of that workers' log will be printed in the end\n",
		},
	),
	Entry(
		"should cancel execution via context cancellation for all workers",
		100*time.Millisecond,
		4,
		parallel.DoTasksOptions{
			MaxNumberOfWorkers: 2,
		},
		newSpyTask(func(ctx context.Context, taskId int) error {
			switch taskId {
			// workers[0] takes tasks[0]
			// workers[1] takes tasks[2]
			case 0, 2:
				<-ctx.Done()
				Expect(ctx.Err()).To(MatchError(context.DeadlineExceeded))

				// order workers with delay based on taskId
				time.Sleep(time.Duration(taskId*100) * time.Millisecond)

				logboek.Context(ctx).LogF("task[%d]: canceled\n", taskId)
				return nil
				// workers[0] won't take tasks[1]
				// workers[0] won't take tasks[3]
			case 1, 3:
				logboek.Context(ctx).LogLn("not printed because parent context is canceled")
				return nil
			default:
				panic(fmt.Sprintf("unexpected taskId: %d", taskId))
			}
		}),
		2,
		MatchError(context.DeadlineExceeded),
		[]string{
			"task[0]: canceled\n",
			"\ntask[2]: canceled\n",
		},
	),
)

type spyTaskFunc struct {
	callsCount atomic.Int32 // prevent race condition
	callback   parallel.TaskFunc
}

func newSpyTask(callback parallel.TaskFunc) *spyTaskFunc {
	return &spyTaskFunc{
		callback: callback,
	}
}

func (s *spyTaskFunc) Callback(ctx context.Context, taskId int) error {
	s.callsCount.Add(1)
	return s.callback(ctx, taskId)
}

func (s *spyTaskFunc) Count() int {
	return int(s.callsCount.Load())
}

type spyOutput struct {
	buf           *bytes.Buffer
	trackedWrites []string
}

func newSpyOutput(trackedCap int) *spyOutput {
	return &spyOutput{
		buf:           bytes.NewBuffer(nil),
		trackedWrites: make([]string, 0, trackedCap),
	}
}

func (s *spyOutput) Write(p []byte) (n int, err error) {
	s.trackedWrites = append(s.trackedWrites, string(p))
	return s.buf.Write(p)
}

func (s *spyOutput) Lines() []string {
	return s.trackedWrites
}
