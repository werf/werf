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
		expectedOutputLinesMatcher types.GomegaMatcher,
	) {
		output := newSpyOutput(numberOfTasks)
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
		Expect(output.Lines()).To(expectedOutputLinesMatcher)
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
		HaveExactElements([]string{}),
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
		Or(
			HaveExactElements([]string{
				"one\n",
				"two\n",
				"\nthree\nfour\n",
				"five\n",
			}),
			HaveExactElements([]string{
				"one\n",
				"two\n", "\nthree\n",
				"four\n",
				"five\n",
			}),
		),
	),
	Entry(
		"should handle error from one of workers (fail fast), "+
			"stop execution via context cancellation for other workers, "+
			"move failed background worker in the end of the printing queue (highlighted), "+
			"resume printing logs starting from non-failed foreground worker up to the last worker in the printing queue: "+
			"[0 {non-failed,foreground,paused,resumed}, 3, 2, 1 {failed,background,highlighted}]",
		time.Duration(0),
		4,
		parallel.DoTasksOptions{
			MaxNumberOfWorkers: 4,
		},
		newSpyTask(func(ctx context.Context, taskId int) error {
			switch taskId {
			case 0:
				logboek.Context(ctx).LogF("workers[%[1]d], task[%[1]d]: is a foreground non-failed worker (1/2)\n", taskId)
				<-ctx.Done()
				Expect(ctx.Err()).To(MatchError(context.Canceled))
				logboek.Context(ctx).LogF("workers[%[1]d], task[%[1]d]: is a foreground non-failed worker (2/2)\n", taskId)
				return nil
			case 1:
				time.Sleep(50 * time.Millisecond)
				logboek.Context(ctx).LogF("workers[%[1]d], task[%[1]d]: is a background failed worker (1/1)\n", taskId)
				return errors.New("task 1 failed")
			case 2:
				time.Sleep(110 * time.Millisecond)
				logboek.Context(ctx).LogF("workers[%[1]d], task[%[1]d]: is a background non-failed worker (1/1)\n", taskId)
				return nil
			case 3:
				time.Sleep(210 * time.Millisecond)
				logboek.Context(ctx).LogF("workers[%[1]d], task[%[1]d]: is a background non-failed worker (1/1)\n", taskId)
				return nil
			default:
				panic(fmt.Sprintf("unexpected taskId: %d", taskId))
			}
		}),
		4,
		MatchError("task 1 failed"),
		Or(
			HaveExactElements([]string{
				"workers[0], task[0]: is a foreground non-failed worker (1/2)\nworkers[0], task[0]: is a foreground non-failed worker (2/2)\n",
				"\nworkers[3], task[3]: is a background non-failed worker (1/1)\n",
				"\nworkers[2], task[2]: is a background non-failed worker (1/1)\n",
				"\nworkers[1], task[1]: is a background failed worker (1/1)\n",
			}),
			HaveExactElements([]string{
				"workers[0], task[0]: is a foreground non-failed worker (1/2)\n",
				"workers[0], task[0]: is a foreground non-failed worker (2/2)\n",
				"\nworkers[3], task[3]: is a background non-failed worker (1/1)\n",
				"\nworkers[2], task[2]: is a background non-failed worker (1/1)\n",
				"\nworkers[1], task[1]: is a background failed worker (1/1)\n",
			}),
		),
	),
	Entry(
		"should handle error from one of workers (fail fast), "+
			"stop execution via context cancellation for other workers, "+
			"resume printing logs of failed foreground worker (highlighted) "+
			"and discard logs from other workers: "+
			"[0 {failed,foreground,paused,resumed,highlighted}, 1 {discarded}, 2 {discarded}, 3 {discarded}]",
		time.Duration(0),
		4,
		parallel.DoTasksOptions{
			MaxNumberOfWorkers: 4,
		},
		newSpyTask(func(ctx context.Context, taskId int) error {
			switch taskId {
			case 0:
				logboek.Context(ctx).LogF("workers[%[1]d], task[%[1]d]: is a foreground failed worker (1/1)\n", taskId)
				time.Sleep(110 * time.Millisecond)
				return errors.New("task 0 failed")
			case 1:
				time.Sleep(50 * time.Millisecond)
				logboek.Context(ctx).LogF("workers[%[1]d], task[%[1]d]: is a background non-failed worker (1/1)\n", taskId)
				return nil
			case 2:
				time.Sleep(110 * time.Millisecond)
				logboek.Context(ctx).LogF("workers[%[1]d], task[%[1]d]: is a background non-failed worker (1/1)\n", taskId)
				return nil
			case 3:
				time.Sleep(210 * time.Millisecond)
				logboek.Context(ctx).LogF("workers[%[1]d], task[%[1]d]: is a background non-failed worker (1/1)\n", taskId)
				return nil
			default:
				panic(fmt.Sprintf("unexpected taskId: %d", taskId))
			}
		}),
		4,
		MatchError("task 0 failed"),
		HaveExactElements([]string{
			"workers[0], task[0]: is a foreground failed worker (1/1)\n",
		}),
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
		HaveExactElements([]string{
			"task[0]: canceled\n",
			"\ntask[2]: canceled\n",
		}),
	),
	Entry(
		"should work with 1 worker per 2 tasks where 1-st task is failed",
		time.Duration(0),
		2,
		parallel.DoTasksOptions{
			MaxNumberOfWorkers: 1,
		},
		newSpyTask(func(ctx context.Context, taskId int) error {
			switch taskId {
			case 0:
				logboek.Context(ctx).LogF("workers[0], task[%[1]d]: is a foreground failed worker (1/2)\n", taskId)
				time.Sleep(110 * time.Millisecond)
				return errors.New("task 0 failed")
			case 1:
				time.Sleep(210 * time.Millisecond)
				logboek.Context(ctx).LogF("workers[0], task[%[1]d]: is a foreground failed worker (2/2)\n", taskId)
				return nil
			default:
				panic(fmt.Sprintf("unexpected taskId: %d", taskId))
			}
		}),
		1,
		MatchError("task 0 failed"),
		HaveExactElements([]string{
			"workers[0], task[0]: is a foreground failed worker (1/2)\n",
		}),
	),
	Entry(
		"should work with 1 worker per 2 tasks where 1-st task is canceled",
		150*time.Millisecond,
		2,
		parallel.DoTasksOptions{
			MaxNumberOfWorkers: 1,
		},
		newSpyTask(func(ctx context.Context, taskId int) error {
			switch taskId {
			case 0:
				logboek.Context(ctx).LogF("workers[0], task[%[1]d]: is a foreground non-failed worker (1/3)\n", taskId)
				<-ctx.Done()
				Expect(ctx.Err()).To(MatchError(context.DeadlineExceeded))
				logboek.Context(ctx).LogF("workers[0], task[%[1]d]: is a foreground non-failed worker (2/3)\n", taskId)
				return nil
			case 1:
				logboek.Context(ctx).LogF("workers[0], task[%[1]d]: is a foreground non-failed worker (3/3)\n", taskId)
				return nil
			default:
				panic(fmt.Sprintf("unexpected taskId: %d", taskId))
			}
		}),
		1,
		MatchError(context.DeadlineExceeded),
		HaveExactElements([]string{
			"workers[0], task[0]: is a foreground non-failed worker (1/3)\n",
			"workers[0], task[0]: is a foreground non-failed worker (2/3)\n",
		}),
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
