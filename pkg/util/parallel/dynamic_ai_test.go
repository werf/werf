package parallel_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

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
