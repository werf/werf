package opstats

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collector pending/commit flush", func() {
	ctx := context.Background()

	It("returns only what was recorded since the last commit", func() {
		c := NewCollector()
		base := time.Now()

		c.add(OperationStageBuild, base, base.Add(time.Second))
		c.add(OperationImagePull, base, base.Add(2*time.Second))
		c.events[EventStageBuilt] = 2

		Expect(c.PendingSummary(ctx)).To(HaveLen(2))
		Expect(c.PendingEventSummary(ctx)).To(Equal([]EventSummary{{Event: EventStageBuilt, Count: 2}}))
		c.CommitFlush(ctx)

		c.add(OperationStageBuild, base.Add(3*time.Second), base.Add(5*time.Second))
		c.events[EventStageBuilt] = 3

		Expect(c.PendingSummary(ctx)).To(Equal([]OperationSummary{{
			Operation: OperationStageBuild,
			Count:     1,
			TotalTime: 2 * time.Second,
			WallTime:  2 * time.Second,
			AvgTime:   2 * time.Second,
			MaxTime:   2 * time.Second,
		}}))
		Expect(c.PendingEventSummary(ctx)).To(Equal([]EventSummary{{Event: EventStageBuilt, Count: 1}}))
		c.CommitFlush(ctx)

		Expect(c.PendingSummary(ctx)).To(BeEmpty())
		Expect(c.PendingEventSummary(ctx)).To(BeEmpty())
	})

	It("retains pending observations until the flush is committed", func() {
		c := NewCollector()
		base := time.Now()

		c.add(OperationStageBuild, base, base.Add(time.Second))
		c.events[EventStageBuilt] = 1

		Expect(c.PendingSummary(ctx)).To(HaveLen(1))
		Expect(c.PendingSummary(ctx)).To(HaveLen(1))
		Expect(c.PendingEventSummary(ctx)).To(HaveLen(1))

		c.add(OperationImagePull, base.Add(time.Second), base.Add(2*time.Second))
		Expect(c.PendingSummary(ctx)).To(HaveLen(2))
	})

	It("keeps Summary and EventSummary cumulative regardless of commits", func() {
		c := NewCollector()
		base := time.Now()

		c.add(OperationStageBuild, base, base.Add(time.Second))
		c.events[EventStageBuilt] = 1
		c.CommitFlush(ctx)
		c.add(OperationStageBuild, base.Add(2*time.Second), base.Add(3*time.Second))
		c.events[EventStageBuilt] = 3

		summary := c.Summary()
		Expect(summary).To(HaveLen(1))
		Expect(summary[0].Count).To(Equal(2))
		Expect(summary[0].TotalTime).To(Equal(2 * time.Second))

		Expect(c.EventSummary()).To(Equal([]EventSummary{{Event: EventStageBuilt, Count: 3}}))
	})
})
