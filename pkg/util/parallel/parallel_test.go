package parallel_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/util/parallel"
)

var _ = DescribeTable("parallel task",
	func(
		ctx context.Context,
		parallelExecutionLimit time.Duration,
		numberOfTasks int,
		options parallel.DoTasksOptions,
		spy *spyTaskFunc,
		expectedCallsCount int,
		errMatcher types.GomegaMatcher,
	) {
		ctx = logging.WithLogger(ctx)

		logboek.Context(ctx).OutStream()

		if parallelExecutionLimit > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, parallelExecutionLimit)
			defer cancel()
		}

		err := parallel.DoTasks(ctx, numberOfTasks, options, spy.Callback)

		Expect(spy.CallsCount).To(Equal(expectedCallsCount))
		Expect(err).To(errMatcher)
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
	),
	// TODO: implement in future
	XEntry(
		"should print log while long task execution",
		time.Duration(0),
		1,
		parallel.DoTasksOptions{
			MaxNumberOfWorkers: 1,
		},
		newSpyTask(func(ctx context.Context, taskId int) error {
			panic("not implemented")
		}),
		0,
		Succeed(),
	),
	Entry(
		"should stop parallel execution process if one of tasks failed (fail fast)",
		time.Duration(0),
		2,
		parallel.DoTasksOptions{
			MaxNumberOfWorkers: 2,
		},
		newSpyTask(func(ctx context.Context, taskId int) error {
			switch taskId {
			case 0:
				select {
				case <-ctx.Done():
					Expect(ctx.Err()).To(MatchError(context.Canceled))
					return nil
				case <-time.After(100 * time.Millisecond):
					return errors.New("task 0 failed")
				}
			case 1:
				time.Sleep(50 * time.Millisecond)
				return errors.New("task 1 failed")
			default:
				panic(fmt.Sprintf("unexpected taskId: %d", taskId))
			}
		}),
		2,
		MatchError("task 1 failed"),
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
		Succeed(),
	),
)

type spyTaskFunc struct {
	CallsCount int

	callback parallel.TaskFunc
}

func newSpyTask(callback parallel.TaskFunc) *spyTaskFunc {
	return &spyTaskFunc{
		callback: callback,
	}
}

func (s *spyTaskFunc) Callback(ctx context.Context, taskId int) error {
	s.CallsCount++
	return s.callback(ctx, taskId)
}
