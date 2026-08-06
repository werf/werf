package opstats

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collector", func() {
	base := time.Now()

	iv := func(startSec, endSec float64) interval {
		return interval{
			start: base.Add(time.Duration(startSec * float64(time.Second))),
			end:   base.Add(time.Duration(endSec * float64(time.Second))),
		}
	}

	DescribeTable("unionDuration",
		func(intervals []interval, expected time.Duration) {
			Expect(unionDuration(intervals)).To(Equal(expected))
		},
		Entry("no intervals", []interval{}, time.Duration(0)),
		Entry("single interval", []interval{iv(0, 10)}, 10*time.Second),
		Entry("non-overlapping intervals", []interval{iv(0, 5), iv(10, 15)}, 10*time.Second),
		Entry("fully overlapping intervals", []interval{iv(0, 10), iv(2, 8)}, 10*time.Second),
		Entry("partially overlapping intervals", []interval{iv(0, 10), iv(5, 15)}, 15*time.Second),
		Entry("unsorted intervals", []interval{iv(10, 15), iv(0, 5), iv(4, 11)}, 15*time.Second),
		Entry("touching intervals", []interval{iv(0, 5), iv(5, 10)}, 10*time.Second),
	)

	It("computes total, avg and max over intervals and sorts by total time", func() {
		collector := NewCollector()

		for _, in := range []interval{iv(0, 10), iv(5, 15), iv(20, 22)} {
			collector.add(OperationImagePush, in.start, in.end)
		}
		collector.add(OperationImagePull, iv(0, 3).start, iv(0, 3).end)

		summary := collector.Summary()
		Expect(summary).To(HaveLen(2))

		push := summary[0]
		Expect(push.Operation).To(Equal(OperationImagePush))
		Expect(push.Count).To(Equal(3))
		Expect(push.TotalTime).To(Equal(22 * time.Second))
		Expect(push.WallTime).To(Equal(17 * time.Second))
		Expect(push.AvgTime).To(Equal(22 * time.Second / 3))
		Expect(push.MaxTime).To(Equal(10 * time.Second))

		pull := summary[1]
		Expect(pull.Operation).To(Equal(OperationImagePull))
		Expect(pull.TotalTime).To(Equal(3 * time.Second))
	})

	It("records at most once when the done func is called twice", func() {
		collector := NewCollector()
		ctx := NewContext(context.Background(), collector)

		done := Observe(ctx, OperationImagePull)
		done()
		done()

		summary := collector.Summary()
		Expect(summary).To(HaveLen(1))
		Expect(summary[0].Count).To(Equal(1))
	})

	It("aggregates operations concurrently and reports union wall time", func() {
		collector := NewCollector()
		ctx := NewContext(context.Background(), collector)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				done := Observe(ctx, OperationImagePull)
				time.Sleep(10 * time.Millisecond)
				done()
			}()
		}
		wg.Wait()

		summary := collector.Summary()
		Expect(summary).To(HaveLen(1))
		Expect(summary[0].Operation).To(Equal(OperationImagePull))
		Expect(summary[0].Count).To(Equal(10))
		Expect(summary[0].WallTime).To(BeNumerically(">=", 10*time.Millisecond))
		Expect(summary[0].WallTime).To(BeNumerically("<", 100*time.Millisecond))
	})

	It("counts events and sorts by count", func() {
		collector := NewCollector()
		ctx := NewContext(context.Background(), collector)

		CountEvent(ctx, EventStageBuilt)
		CountEvent(ctx, EventStageCacheHitLocal)
		CountEvent(ctx, EventStageCacheHitLocal)

		events := collector.EventSummary()
		Expect(events).To(Equal([]EventSummary{
			{Event: EventStageCacheHitLocal, Count: 2},
			{Event: EventStageBuilt, Count: 1},
		}))
	})

	It("is a no-op without collector in context", func() {
		done := Observe(context.Background(), OperationImagePull)
		Expect(done).NotTo(BeNil())
		done()

		CountEvent(context.Background(), EventStageBuilt)
	})
})
