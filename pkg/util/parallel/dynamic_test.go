package parallel_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/util/parallel"
	"github.com/werf/werf/v2/pkg/werf"
)

var _ = Describe("DoTasksDynamic", func() {
	BeforeEach(func() {
		// tmp_manager requires werf init
		Expect(werf.Init(GinkgoT().TempDir(), "")).To(Succeed())
	})

	It("runs every dynamically produced task exactly once, respecting the worker cap", func() {
		const total = 8
		const maxWorkers = 3

		var mu sync.Mutex
		produced := 0
		next := func(ctx context.Context) (int, bool, error) {
			mu.Lock()
			defer mu.Unlock()
			if produced >= total {
				return 0, false, nil
			}
			id := produced
			produced++
			return id, true, nil
		}

		var inFlight, maxInFlight atomic.Int32
		var callsCount atomic.Int32
		seen := make([]atomic.Bool, total)

		err := parallel.DoTasksDynamic(context.Background(), parallel.DoTasksOptions{MaxNumberOfWorkers: maxWorkers}, next, func(ctx context.Context, taskId int) error {
			cur := inFlight.Add(1)
			defer inFlight.Add(-1)
			for {
				m := maxInFlight.Load()
				if cur <= m || maxInFlight.CompareAndSwap(m, cur) {
					break
				}
			}

			callsCount.Add(1)
			Expect(seen[taskId].Swap(true)).To(BeFalse(), "task %d must run exactly once", taskId)

			return nil
		})

		Expect(err).To(Succeed())
		Expect(callsCount.Load()).To(Equal(int32(total)))
		Expect(maxInFlight.Load()).To(BeNumerically("<=", int32(maxWorkers)))
	})

	It("actually runs tasks concurrently, not just below a cap", func() {
		// Deterministic proof of real concurrency (not a timing-based flaky
		// check): every one of `workers` tasks must reach the barrier before
		// any of them is allowed to return. A sequential (or otherwise
		// broken) scheduler would deadlock here — task 0 would sit on the
		// barrier forever because no other task ever gets to start — and
		// the test fails via the context timeout instead of hanging.
		const workers = 3

		var mu sync.Mutex
		produced := 0
		next := func(ctx context.Context) (int, bool, error) {
			mu.Lock()
			defer mu.Unlock()
			if produced >= workers {
				return 0, false, nil
			}
			id := produced
			produced++
			return id, true, nil
		}

		var arrived atomic.Int32
		barrier := make(chan struct{})
		var closeOnce sync.Once

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := parallel.DoTasksDynamic(ctx, parallel.DoTasksOptions{MaxNumberOfWorkers: workers}, next, func(ctx context.Context, taskId int) error {
			if arrived.Add(1) == int32(workers) {
				closeOnce.Do(func() { close(barrier) })
			}

			select {
			case <-barrier:
				return nil
			case <-ctx.Done():
				return fmt.Errorf("task %d timed out waiting for the other %d worker(s) to start concurrently — DoTasksDynamic is not running tasks in parallel", taskId, workers-1)
			}
		})

		Expect(err).To(Succeed())
	})

	It("fails fast: an error from one task stops scheduling of further tasks", func() {
		var mu sync.Mutex
		produced := 0

		next := func(ctx context.Context) (int, bool, error) {
			select {
			case <-ctx.Done():
				return 0, false, ctx.Err()
			default:
			}

			mu.Lock()
			defer mu.Unlock()
			id := produced
			produced++
			return id, true, nil
		}

		err := parallel.DoTasksDynamic(context.Background(), parallel.DoTasksOptions{MaxNumberOfWorkers: 2}, next, func(ctx context.Context, taskId int) error {
			if taskId == 0 {
				return errors.New("boom")
			}

			<-ctx.Done()
			return nil
		})

		Expect(err).To(MatchError(ContainSubstring("boom")))
	})

	It("returns immediately when there are no tasks to run", func() {
		next := func(ctx context.Context) (int, bool, error) {
			return 0, false, nil
		}

		called := false
		err := parallel.DoTasksDynamic(context.Background(), parallel.DoTasksOptions{MaxNumberOfWorkers: 2}, next, func(ctx context.Context, taskId int) error {
			called = true
			return nil
		})

		Expect(err).To(Succeed())
		Expect(called).To(BeFalse())
	})

	It("streams a running task's output while another worker idles waiting for a dependent task", func() {
		// Worker 0 gets the short task A and then blocks in next() until the
		// long task B on worker 1 completes, because C depends on B. B does
		// not finish until it sees its own first line reach the sink — with
		// output printed per worker rather than per task, the printer would
		// stay on the idle worker 0, B's line would never be flushed and the
		// run would deadlock until the context deadline. Worker 2 never gets
		// a task at all and must not leave a hole in the start order.
		const (
			taskA = 0
			taskB = 1
			taskC = 2
		)

		sink := newSpyOutput(8)
		ctx := logboek.NewContext(context.Background(), logboek.NewLogger(sink, sink))
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var mu sync.Mutex
		var aGiven, bGiven, cGiven bool
		bDone := make(chan struct{})
		startOrder := map[int]int{}
		var enqueued []int

		next := func(ctx context.Context) (int, bool, error) {
			workerID := ctx.Value(parallel.CtxBackgroundTaskIDKey).(int)

			mu.Lock()
			switch {
			case workerID == 0 && !aGiven:
				aGiven = true
				mu.Unlock()
				return taskA, true, nil
			case workerID == 1 && !bGiven:
				bGiven = true
				mu.Unlock()
				return taskB, true, nil
			}
			mu.Unlock()

			select {
			case <-bDone:
			case <-ctx.Done():
				return 0, false, ctx.Err()
			}

			mu.Lock()
			defer mu.Unlock()
			if cGiven {
				return 0, false, nil
			}
			cGiven = true
			return taskC, true, nil
		}

		err := parallel.DoTasksDynamic(ctx, parallel.DoTasksOptions{
			MaxNumberOfWorkers: 3,
			OnTaskEnqueued: func(taskID, startOrder int) {
				Expect(startOrder).To(Equal(len(enqueued)))
				enqueued = append(enqueued, taskID)
			},
		}, next, func(ctx context.Context, taskId int) error {
			order, ok := parallel.TaskStartOrder(ctx)
			Expect(ok).To(BeTrue(), "a task context always carries its start order")

			mu.Lock()
			startOrder[taskId] = order
			mu.Unlock()

			switch taskId {
			case taskA:
				logboek.Context(ctx).LogLn("a")
				return nil
			case taskB:
				defer close(bDone)
				logboek.Context(ctx).LogLn("b-start")
				for !strings.Contains(sink.String(), "b-start") {
					select {
					case <-ctx.Done():
						return fmt.Errorf("task B: its first line never reached the sink while it was running: %w", ctx.Err())
					case <-time.After(10 * time.Millisecond):
					}
				}
				logboek.Context(ctx).LogLn("b-end")
				return nil
			case taskC:
				logboek.Context(ctx).LogLn("c")
				return nil
			default:
				return fmt.Errorf("unexpected task %d", taskId)
			}
		})

		Expect(err).To(Succeed())
		Expect(sink.String()).To(Equal("a\n\nb-start\nb-end\n\nc\n"))
		Expect(startOrder).To(Equal(map[int]int{taskA: 0, taskB: 1, taskC: 2}))
		Expect(enqueued).To(Equal([]int{taskA, taskB, taskC}))
	})

	It("leaves the printing queue in start order when the error comes from next() and not from a task", func() {
		// Worker 1 finishes its task successfully and only then fails while
		// asking for the next one. No block is to blame for that error, so
		// none may be pulled out of the queue: the blocks must still read
		// a, b, c. Worker 0 holds the head of the queue for the whole run, so
		// the printer is still parked on it when the error arrives.
		sink := newSpyOutput(8)
		ctx := logboek.NewContext(context.Background(), logboek.NewLogger(sink, sink))
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var mu sync.Mutex
		given := map[int]bool{}
		cStarted := make(chan struct{})

		next := func(ctx context.Context) (int, bool, error) {
			workerID := ctx.Value(parallel.CtxBackgroundTaskIDKey).(int)

			mu.Lock()
			first := !given[workerID]
			given[workerID] = true
			mu.Unlock()

			if first {
				return workerID, true, nil
			}

			if workerID == 1 {
				select {
				case <-cStarted:
					return 0, false, errors.New("scheduler failed")
				case <-ctx.Done():
					return 0, false, ctx.Err()
				}
			}

			<-ctx.Done()
			return 0, false, ctx.Err()
		}

		err := parallel.DoTasksDynamic(ctx, parallel.DoTasksOptions{MaxNumberOfWorkers: 3}, next, func(ctx context.Context, taskId int) error {
			switch taskId {
			case 0:
				logboek.Context(ctx).LogLn("a")
				<-ctx.Done()
				return nil
			case 1:
				logboek.Context(ctx).LogLn("b")
				return nil
			case 2:
				logboek.Context(ctx).LogLn("c")
				close(cStarted)
				return nil
			default:
				return fmt.Errorf("unexpected task %d", taskId)
			}
		})

		Expect(err).To(MatchError(ContainSubstring("scheduler failed")))
		Expect(sink.String()).To(Equal("a\n\nb\n\nc\n"))
	})

	It("runs tasks sequentially on a single worker when MaxNumberOfWorkers is not positive", func() {
		const total = 4

		var mu sync.Mutex
		produced := 0
		next := func(ctx context.Context) (int, bool, error) {
			mu.Lock()
			defer mu.Unlock()
			if produced >= total {
				return 0, false, nil
			}
			id := produced
			produced++
			return id, true, nil
		}

		var inFlight, maxInFlight atomic.Int32
		var tasksRun []int

		err := parallel.DoTasksDynamic(context.Background(), parallel.DoTasksOptions{MaxNumberOfWorkers: 0}, next, func(ctx context.Context, taskId int) error {
			cur := inFlight.Add(1)
			defer inFlight.Add(-1)
			for {
				m := maxInFlight.Load()
				if cur <= m || maxInFlight.CompareAndSwap(m, cur) {
					break
				}
			}

			tasksRun = append(tasksRun, taskId)
			return nil
		})

		Expect(err).To(Succeed())
		Expect(tasksRun).To(Equal([]int{0, 1, 2, 3}))
		Expect(maxInFlight.Load()).To(Equal(int32(1)))
	})
})
