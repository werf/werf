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
				logboek.Context(ctx).LogLn("one")
				time.Sleep(300 * time.Millisecond)
			case 1:
				logboek.Context(ctx).LogLn("two")
				time.Sleep(200 * time.Millisecond)
			case 2:
				logboek.Context(ctx).LogLn("three")
				time.Sleep(100 * time.Millisecond)
			case 3:
				logboek.Context(ctx).LogLn("four")
			}
			return nil
		}),
		4,
		Succeed(),
		[]string{
			"one\n",
			"two\n",
			"\nthree\nfour\n",
		},
	),
	Entry(
		"should stop parallel execution process if one of tasks failed (fail fast) and print log from failed task",
		time.Duration(0),
		2,
		parallel.DoTasksOptions{
			MaxNumberOfWorkers: 2,
		},
		newSpyTask(func(ctx context.Context, taskId int) error {
			switch taskId {
			case 0:
				<-ctx.Done()
				Expect(ctx.Err()).To(MatchError(context.Canceled))
				logboek.Context(ctx).LogLn("task 0 must not print log")
				return nil
			case 1:
				logboek.Context(ctx).LogLn("task 1 prints log")
				return errors.New("task 1 failed")
			default:
				panic(fmt.Sprintf("unexpected taskId: %d", taskId))
			}
		}),
		2,
		MatchError("task 1 failed"),
		[]string{
			"task 1 prints log\n",
		},
	),
	Entry(
		"should cancel parallel execution if parent context is canceled",
		100*time.Millisecond,
		2,
		parallel.DoTasksOptions{
			MaxNumberOfWorkers: 2,
		},
		newSpyTask(func(ctx context.Context, taskId int) error {
			select {
			case <-ctx.Done():
				Expect(ctx.Err()).To(MatchError(context.DeadlineExceeded))
				return nil
			case <-time.After(1 * time.Second):
				return errors.New("task execution timeout")
			}
		}),
		2,
		MatchError(context.DeadlineExceeded),
		[]string{},
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
