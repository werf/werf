package opstats

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collector flush", func() {
	It("returns only what was recorded since the previous flush", func() {
		c := NewCollector()
		base := time.Now()

		c.add(OperationStageBuild, base, base.Add(time.Second))
		c.add(OperationImagePull, base, base.Add(2*time.Second))
		c.events[EventStageBuilt] = 2

		first := c.FlushSummary()
		Expect(first).To(HaveLen(2))
		Expect(c.FlushEventSummary()).To(Equal([]EventSummary{{Event: EventStageBuilt, Count: 2}}))

		c.add(OperationStageBuild, base.Add(3*time.Second), base.Add(5*time.Second))
		c.events[EventStageBuilt] = 3

		second := c.FlushSummary()
		Expect(second).To(Equal([]OperationSummary{{
			Operation: OperationStageBuild,
			Count:     1,
			TotalTime: 2 * time.Second,
			WallTime:  2 * time.Second,
			AvgTime:   2 * time.Second,
			MaxTime:   2 * time.Second,
		}}))
		Expect(c.FlushEventSummary()).To(Equal([]EventSummary{{Event: EventStageBuilt, Count: 1}}))

		Expect(c.FlushSummary()).To(BeEmpty())
		Expect(c.FlushEventSummary()).To(BeEmpty())
	})

	It("keeps Summary cumulative regardless of flushes", func() {
		c := NewCollector()
		base := time.Now()

		c.add(OperationStageBuild, base, base.Add(time.Second))
		c.FlushSummary()
		c.add(OperationStageBuild, base.Add(2*time.Second), base.Add(3*time.Second))

		summary := c.Summary()
		Expect(summary).To(HaveLen(1))
		Expect(summary[0].Count).To(Equal(2))
		Expect(summary[0].TotalTime).To(Equal(2 * time.Second))
	})
})
