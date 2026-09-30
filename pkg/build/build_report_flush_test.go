package build

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/opstats"
)

var _ = Describe("createBuildReport operations flush", func() {
	It("does not lose pending observations when the report write fails", func() {
		collector := opstats.NewCollector()
		ctx := opstats.NewContext(context.Background(), collector)

		opstats.Observe(ctx, opstats.OperationStageBuild)()
		opstats.CountEvent(ctx, opstats.EventStageBuilt)
		opstats.CountEvent(ctx, opstats.EventRegistryTagsCacheHit)

		failPhase := newReportPhase(filepath.Join(GinkgoT().TempDir(), "missing-dir", "report.json"))
		Expect(createBuildReport(ctx, failPhase, nil)).To(HaveOccurred())

		opstats.Observe(ctx, opstats.OperationStageBuild)()
		opstats.CountEvent(ctx, opstats.EventStageBuilt)
		opstats.CountEvent(ctx, opstats.EventRegistryTagsCacheHit)

		reportPath := filepath.Join(GinkgoT().TempDir(), "report.json")
		Expect(createBuildReport(ctx, newReportPhase(reportPath), nil)).To(Succeed())

		decoded := readOperationsReport(reportPath)
		Expect(decoded.Operations[string(opstats.OperationStageBuild)].Count).To(Equal(2))
		Expect(decoded.StageCache).To(Equal(map[string]int{string(opstats.EventStageBuilt): 2}))
		Expect(decoded.RegistryCache).To(Equal(map[string]int{string(opstats.EventRegistryTagsCacheHit): 2}))

		Expect(collector.PendingSummary(ctx)).To(BeEmpty())
		Expect(collector.PendingEventSummary(ctx)).To(BeEmpty())
	})

	It("reports only the counters recorded since the previous report", func() {
		collector := opstats.NewCollector()
		ctx := opstats.NewContext(context.Background(), collector)

		opstats.CountEvent(ctx, opstats.EventStageBuilt)
		opstats.CountEvent(ctx, opstats.EventRegistryTagsCacheHit)
		Expect(createBuildReport(ctx, newReportPhase(filepath.Join(GinkgoT().TempDir(), "first.json")), nil)).To(Succeed())

		opstats.CountEvent(ctx, opstats.EventRegistryTagsSharedResult)
		secondPath := filepath.Join(GinkgoT().TempDir(), "second.json")
		Expect(createBuildReport(ctx, newReportPhase(secondPath), nil)).To(Succeed())

		decoded := readOperationsReport(secondPath)
		Expect(decoded.StageCache).To(BeEmpty())
		Expect(decoded.RegistryCache).To(Equal(map[string]int{string(opstats.EventRegistryTagsSharedResult): 1}))

		Expect(collector.EventSummary()).To(ConsistOf(
			opstats.EventSummary{Event: opstats.EventStageBuilt, Count: 1},
			opstats.EventSummary{Event: opstats.EventRegistryTagsCacheHit, Count: 1},
			opstats.EventSummary{Event: opstats.EventRegistryTagsSharedResult, Count: 1},
		))
	})
})
