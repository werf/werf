package opstats

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("LogSummary", func() {
	base := time.Now()

	ginkgo.It("renders the operations table with its two legend lines and no command time", func() {
		collector := NewCollector()
		collector.add(OperationRegistryTagsList, base, base.Add(2*time.Second))
		collector.add(OperationRegistryTagsList, base.Add(time.Second), base.Add(3*time.Second))

		output := logSummaryOutput(collector)

		gomega.Expect(output).To(gomega.MatchRegexp(`Operations`))
		gomega.Expect(output).To(gomega.MatchRegexp(`operation\s+count\s+sum\s+union\s+avg\s+max`))
		gomega.Expect(output).To(gomega.MatchRegexp(`registry: tags list\s+2\s+4\.00s\s+3\.00s\s+2\.00s\s+2\.00s\s+×1\.3 \(sum/union\)`))
		gomega.Expect(output).To(gomega.ContainSubstring("sum: durations added; parallel calls counted separately\n"))
		gomega.Expect(output).To(gomega.ContainSubstring("union: time with ≥1 active call; overlapping intervals counted once\n"))
		gomega.Expect(output).NotTo(gomega.ContainSubstring("build time"))
		gomega.Expect(output).NotTo(gomega.ContainSubstring("command time"))
	})

	ginkgo.It("keeps every column aligned for the longest operation name werf emits", func() {
		collector := NewCollector()
		// 31 characters, the longest label any Observe call site produces.
		collector.add(Operation("registry: image mutate and push"), base, base.Add(2*time.Second))
		collector.add(OperationGitClone, base, base.Add(time.Second))

		output := logSummaryOutput(collector)

		gomega.Expect(output).To(gomega.ContainSubstring(
			"operation                            count       sum     union       avg       max\n"))
		gomega.Expect(output).To(gomega.ContainSubstring(
			"registry: image mutate and push          1     2.00s     2.00s     2.00s     2.00s\n"))
		gomega.Expect(output).To(gomega.ContainSubstring(
			"git: clone                               1     1.00s     1.00s     1.00s     1.00s\n"))
	})

	ginkgo.It("keeps the cache table aligned with the operations table, undefined hit rate included", func() {
		collector := NewCollector()
		ctx := NewContext(context.Background(), collector)
		for range 3 {
			CountCacheLookup(ctx, Operation("registry: image mutate and push"), CacheOutcomeBypass, false)
		}

		output := logSummaryOutput(collector)

		gomega.Expect(output).To(gomega.ContainSubstring(
			"operation                       lookups    hit   miss bypass shared   hit%\n"))
		// The undefined mark is one rune, and fmt pads by runes, so it needs no padding of its own.
		gomega.Expect(output).To(gomega.ContainSubstring(
			"registry: image mutate and push       3      0      0      3      0      —\n"))
	})

	ginkgo.It("omits the parallelism column for sequential calls", func() {
		collector := NewCollector()
		collector.add(OperationRegistryTagsList, base, base.Add(time.Second))

		gomega.Expect(logSummaryOutput(collector)).NotTo(gomega.ContainSubstring("sum/union"))
	})

	ginkgo.It("renders the cache table with its three legend lines", func() {
		collector := NewCollector()
		ctx := NewContext(context.Background(), collector)
		CountCacheLookup(ctx, OperationRegistryTagsList, CacheOutcomeHit, false)
		CountCacheLookup(ctx, OperationRegistryTagsList, CacheOutcomeMiss, false)
		CountCacheLookup(ctx, OperationRegistryTagsList, CacheOutcomeMiss, true)
		CountCacheLookup(ctx, OperationRegistryTagsList, CacheOutcomeBypass, false)

		output := logSummaryOutput(collector)

		gomega.Expect(output).To(gomega.MatchRegexp(`Cache summary`))
		gomega.Expect(output).To(gomega.MatchRegexp(`operation\s+lookups\s+hit\s+miss\s+bypass\s+shared\s+hit%`))
		gomega.Expect(output).To(gomega.MatchRegexp(`registry: tags list\s+4\s+1\s+2\s+1\s+1\s+33%`))
		gomega.Expect(output).To(gomega.ContainSubstring("lookups = hit + miss + bypass\n"))
		gomega.Expect(output).To(gomega.ContainSubstring("hit% = hit / (hit + miss); bypass excluded\n"))
		gomega.Expect(output).To(gomega.ContainSubstring("shared: calls joining an in-flight request; included in miss or bypass\n"))
	})

	ginkgo.DescribeTable("renders the hit rate",
		func(hit, miss, bypass int, expected string) {
			collector := NewCollector()
			ctx := NewContext(context.Background(), collector)
			for range hit {
				CountCacheLookup(ctx, OperationDockerImageList, CacheOutcomeHit, false)
			}
			for range miss {
				CountCacheLookup(ctx, OperationDockerImageList, CacheOutcomeMiss, false)
			}
			for range bypass {
				CountCacheLookup(ctx, OperationDockerImageList, CacheOutcomeBypass, false)
			}

			gomega.Expect(logSummaryOutput(collector)).To(gomega.MatchRegexp(`docker: image list\s+\d+\s+\d+\s+\d+\s+\d+\s+0\s+` + expected))
		},
		ginkgo.Entry("as zero when the cache answered nothing it was asked", 0, 2, 0, `0%`),
		ginkgo.Entry("as a share of the lookups that consulted the cache", 1, 3, 0, `25%`),
		ginkgo.Entry("as undefined when every lookup bypassed the cache", 0, 0, 3, `—`),
	)

	ginkgo.DescribeTable("renders the stages line",
		func(events map[Event]int, expected, forbidden []string) {
			collector := NewCollector()
			ctx := NewContext(context.Background(), collector)
			for event, count := range events {
				for range count {
					CountEvent(ctx, event)
				}
			}

			output := logSummaryOutput(collector)
			for _, pattern := range expected {
				gomega.Expect(output).To(gomega.MatchRegexp(pattern))
			}
			for _, pattern := range forbidden {
				gomega.Expect(output).NotTo(gomega.MatchRegexp(pattern))
			}
		},
		ginkgo.Entry("summing every reuse source",
			map[Event]int{EventStageCacheHitLocal: 4, EventStageCacheHitRepo: 2, EventStageCacheHitSecondary: 1, EventStageBuilt: 3},
			[]string{`Stages: 7 reused, 3 built, 0 discarded`},
			[]string{`Stage cache summary`, `Registry cache summary`, `Recovery:`},
		),
		ginkgo.Entry("with zero built for a fully reused build",
			map[Event]int{EventStageCacheHitLocal: 2},
			[]string{`Stages: 2 reused, 0 built, 0 discarded`},
			nil,
		),
		ginkgo.Entry("counting a lost publication race as discarded inside reused",
			map[Event]int{EventStageCacheHitRepo: 2, EventStageDiscarded: 1, EventStageBuilt: 1},
			[]string{
				`Stages: 2 reused, 1 built, 1 discarded`,
				`discarded: built locally but another published stage was reused; included in reused`,
			},
			nil,
		),
		ginkgo.Entry("recovery line next to the stages line",
			map[Event]int{EventStageBuilt: 1, EventStageBroken: 2, EventConveyorRestart: 1},
			[]string{`Stages: 0 reused, 1 built, 0 discarded`, `Recovery: 2 broken stage detections, 1 conveyor restarts`},
			nil,
		),
		ginkgo.Entry("recovery line alone when no stage was worked on",
			map[Event]int{EventConveyorRestart: 3},
			[]string{`Recovery: 0 broken stage detections, 3 conveyor restarts`},
			[]string{`Stages:`},
		),
		ginkgo.Entry("not at all when the registry events are the only ones",
			map[Event]int{EventRegistryTagsCacheHit: 3, EventRegistryTagsSharedResult: 2},
			nil,
			[]string{`Stages:`, `Recovery:`, `registry tags cache hit`},
		),
		ginkgo.Entry("not at all without stage events",
			map[Event]int{},
			nil,
			[]string{`Stages:`, `Recovery:`},
		),
	)

	ginkgo.It("prints nothing for an empty collector", func() {
		gomega.Expect(logSummaryOutput(NewCollector())).To(gomega.BeEmpty())
	})

	ginkgo.It("prints nothing without a collector", func() {
		gomega.Expect(logSummaryOutput(nil)).To(gomega.BeEmpty())
	})
})
