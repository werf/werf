package parallel_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

		err := parallel.DoTasksDynamic(context.Background(), maxWorkers, parallel.DoTasksOptions{MaxNumberOfWorkers: maxWorkers}, next, func(ctx context.Context, taskId int) error {
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

		err := parallel.DoTasksDynamic(ctx, workers, parallel.DoTasksOptions{MaxNumberOfWorkers: workers}, next, func(ctx context.Context, taskId int) error {
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

		err := parallel.DoTasksDynamic(context.Background(), 2, parallel.DoTasksOptions{MaxNumberOfWorkers: 2}, next, func(ctx context.Context, taskId int) error {
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
		err := parallel.DoTasksDynamic(context.Background(), 2, parallel.DoTasksOptions{MaxNumberOfWorkers: 2}, next, func(ctx context.Context, taskId int) error {
			called = true
			return nil
		})

		Expect(err).To(Succeed())
		Expect(called).To(BeFalse())
	})
})
