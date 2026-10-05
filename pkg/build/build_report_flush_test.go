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
		opstats.CountCacheLookup(ctx, opstats.OperationRegistryTagsList, opstats.CacheLayerMemory, opstats.CacheOutcomeHit, false)

		failPhase := newReportPhase(filepath.Join(GinkgoT().TempDir(), "missing-dir", "report.json"))
		Expect(createBuildReport(ctx, failPhase, nil)).To(HaveOccurred())

		opstats.Observe(ctx, opstats.OperationStageBuild)()
		opstats.CountEvent(ctx, opstats.EventStageBuilt)
		opstats.CountEvent(ctx, opstats.EventRegistryTagsCacheHit)
		opstats.CountCacheLookup(ctx, opstats.OperationRegistryTagsList, opstats.CacheLayerMemory, opstats.CacheOutcomeMiss, true)

		reportPath := filepath.Join(GinkgoT().TempDir(), "report.json")
		Expect(createBuildReport(ctx, newReportPhase(reportPath), nil)).To(Succeed())

		decoded := readOperationsReport(reportPath)
		Expect(decoded.Operations[string(opstats.OperationStageBuild)].Count).To(Equal(2))
		Expect(decoded.StageCache).To(Equal(map[string]int{string(opstats.EventStageBuilt): 2}))
		Expect(decoded.RegistryCache).To(Equal(map[string]int{string(opstats.EventRegistryTagsCacheHit): 2}))
		Expect(decoded.CacheOperations).To(Equal(map[string]map[string]ReportCacheOperationRecord{
			string(opstats.OperationRegistryTagsList): {"memory": {Lookups: 2, Hit: 1, Miss: 1, Shared: 1}},
		}))

		Expect(collector.PendingSummary(ctx)).To(BeEmpty())
		Expect(collector.PendingEventSummary(ctx)).To(BeEmpty())
		Expect(collector.PendingCacheSummary(ctx)).To(BeEmpty())
	})

	It("reports only the cache lookups recorded since the previous report", func() {
		collector := opstats.NewCollector()
		ctx := opstats.NewContext(context.Background(), collector)

		opstats.CountCacheLookup(ctx, opstats.OperationRegistryTagsList, opstats.CacheLayerMemory, opstats.CacheOutcomeHit, false)
		Expect(createBuildReport(ctx, newReportPhase(filepath.Join(GinkgoT().TempDir(), "first.json")), nil)).To(Succeed())

		opstats.CountCacheLookup(ctx, opstats.OperationDockerImageList, opstats.CacheLayerMemory, opstats.CacheOutcomeBypass, false)
		secondPath := filepath.Join(GinkgoT().TempDir(), "second.json")
		Expect(createBuildReport(ctx, newReportPhase(secondPath), nil)).To(Succeed())

		Expect(readOperationsReport(secondPath).CacheOperations).To(Equal(map[string]map[string]ReportCacheOperationRecord{
			string(opstats.OperationDockerImageList): {"memory": {Lookups: 1, Bypass: 1}},
		}))
	})

	It("flushes the layers of one operation independently of each other", func() {
		collector := opstats.NewCollector()
		ctx := opstats.NewContext(context.Background(), collector)

		opstats.CountCacheLookup(ctx, opstats.OperationGitPatch, opstats.CacheLayerMemory, opstats.CacheOutcomeMiss, false)
		opstats.CountCacheLookup(ctx, opstats.OperationGitPatch, opstats.CacheLayerDisk, opstats.CacheOutcomeHit, false)

		firstPath := filepath.Join(GinkgoT().TempDir(), "first.json")
		Expect(createBuildReport(ctx, newReportPhase(firstPath), nil)).To(Succeed())

		// One lookup of a layered cache is one lookup on each layer it reached, so the
		// two layers of the same operation are two entries under the same key.
		Expect(readOperationsReport(firstPath).CacheOperations).To(Equal(map[string]map[string]ReportCacheOperationRecord{
			string(opstats.OperationGitPatch): {
				"memory": {Lookups: 1, Miss: 1},
				"disk":   {Lookups: 1, Hit: 1},
			},
		}))

		opstats.CountCacheLookup(ctx, opstats.OperationGitPatch, opstats.CacheLayerMemory, opstats.CacheOutcomeHit, false)
		opstats.CountCacheLookup(ctx, opstats.OperationGitChecksum, opstats.CacheLayerDisk, opstats.CacheOutcomeMiss, false)

		secondPath := filepath.Join(GinkgoT().TempDir(), "second.json")
		Expect(createBuildReport(ctx, newReportPhase(secondPath), nil)).To(Succeed())

		// Committing the first report advanced the checkpoint of each layer on its own:
		// the new memory lookup is reported without resurrecting the flushed disk one,
		// and a layer seen for the first time starts from its own zero.
		Expect(readOperationsReport(secondPath).CacheOperations).To(Equal(map[string]map[string]ReportCacheOperationRecord{
			string(opstats.OperationGitPatch):    {"memory": {Lookups: 1, Hit: 1}},
			string(opstats.OperationGitChecksum): {"disk": {Lookups: 1, Miss: 1}},
		}))

		Expect(collector.CacheSummary(ctx)).To(Equal([]opstats.CacheSummary{
			{Operation: opstats.OperationGitChecksum, Layer: opstats.CacheLayerDisk, Miss: 1},
			{Operation: opstats.OperationGitPatch, Layer: opstats.CacheLayerMemory, Hit: 1, Miss: 1},
			{Operation: opstats.OperationGitPatch, Layer: opstats.CacheLayerDisk, Hit: 1},
		}))
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
