package opstats

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.DescribeTable("LogSummary cache blocks",
	func(events map[Event]int, expected, forbidden []string) {
		collector := NewCollector()
		ctx := NewContext(context.Background(), collector)
		Observe(ctx, OperationStageBuild)()
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
	ginkgo.Entry("stage and registry events",
		map[Event]int{EventStageBuilt: 1, EventStageCacheHitRepo: 2, EventRegistryTagsCacheHit: 3, EventRegistryTagsSharedResult: 2},
		[]string{
			`(?s)Operations summary.*Stage cache summary.*Registry cache summary`,
			`found in repo stages storage\s+2 stage\(s\)`,
			`built\s+1 stage\(s\)`,
			`total: 3 stage\(s\)`,
			`registry tags cache hit\s+3 request\(s\)`,
			`registry tags shared result\s+2 request\(s\)`,
			`total: 5 request\(s\)`,
		},
		[]string{`registry tags.*stage\(s\)`},
	),
	ginkgo.Entry("registry events only",
		map[Event]int{EventRegistryTagsCacheHit: 3},
		[]string{`Registry cache summary`, `registry tags cache hit\s+3 request\(s\)`, `total: 3 request\(s\)`},
		[]string{`Stage cache summary`, `stage\(s\)`},
	),
	ginkgo.Entry("stage events only",
		map[Event]int{EventStageBuilt: 1, EventStageCacheHitLocal: 4, EventStageCacheHitSecondary: 1},
		[]string{`Stage cache summary`, `copied from secondary storage\s+1 stage\(s\)`, `total: 6 stage\(s\)`},
		[]string{`Registry cache summary`, `request\(s\)`},
	),
	ginkgo.Entry("no events",
		map[Event]int{},
		[]string{`Operations summary`},
		[]string{`cache summary`, `stage\(s\)`, `request\(s\)`},
	),
)
