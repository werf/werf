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

	It("is a no-op without collector in context", func() {
		done := Observe(context.Background(), OperationImagePull)
		Expect(done).NotTo(BeNil())
		done()
	})
})
