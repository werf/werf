package build

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/opstats"
)

var _ = Describe("ImagesReport operations summary", func() {
	ctx := context.Background()

	It("serializes operations and splits stage and registry cache counters to json", func() {
		report := NewImagesReport()
		report.SetOperationsSummary(ctx,
			[]opstats.OperationSummary{
				{
					Operation: opstats.OperationImagePush,
					Count:     2,
					TotalTime: 3 * time.Second,
					WallTime:  2 * time.Second,
					AvgTime:   1500 * time.Millisecond,
					MaxTime:   2 * time.Second,
				},
			},
			[]opstats.EventSummary{
				{Event: opstats.EventRegistryTagsCacheHit, Count: 3},
				{Event: opstats.EventStageCacheHitRepo, Count: 2},
				{Event: opstats.EventStageBuilt, Count: 1},
				{Event: opstats.EventStageDiscarded, Count: 1},
				{Event: opstats.EventStageBroken, Count: 2},
				{Event: opstats.EventConveyorRestart, Count: 1},
				{Event: opstats.EventRegistryTagsSharedResult, Count: 1},
			},
		)

		data, err := report.ToJsonData()
		Expect(err).NotTo(HaveOccurred())

		decoded := decodeOperationsReport(data)
		Expect(decoded.Operations).To(Equal(map[string]ReportOperationRecord{
			"image push": {
				Count:            2,
				TotalTimeSeconds: 3,
				WallTimeSeconds:  2,
				AvgTimeSeconds:   1.5,
				MaxTimeSeconds:   2,
			},
		}))
		Expect(decoded.StageCache).To(Equal(map[string]int{"built": 1, "found in repo stages storage": 2, "discarded": 1}))
		Expect(decoded.RegistryCache).To(Equal(map[string]int{"registry tags cache hit": 3, "registry tags shared result": 1}))
		Expect(decoded.Recovery).To(Equal(map[string]int{"broken stage detections": 2, "conveyor restarts": 1}))
	})

	DescribeTable("omits a cache section that has no events",
		func(events []opstats.EventSummary, stageCachePresent, registryCachePresent, recoveryPresent bool) {
			report := NewImagesReport()
			report.SetOperationsSummary(ctx, nil, events)

			data, err := report.ToJsonData()
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.Contains(string(data), `"StageCache"`)).To(Equal(stageCachePresent))
			Expect(strings.Contains(string(data), `"RegistryCache"`)).To(Equal(registryCachePresent))
			Expect(strings.Contains(string(data), `"Recovery"`)).To(Equal(recoveryPresent))
		},
		Entry("stage events only", []opstats.EventSummary{{Event: opstats.EventStageBuilt, Count: 1}}, true, false, false),
		Entry("registry events only", []opstats.EventSummary{{Event: opstats.EventRegistryTagsCacheHit, Count: 1}}, false, true, false),
		Entry("recovery events only", []opstats.EventSummary{{Event: opstats.EventStageBroken, Count: 1}}, false, false, true),
		Entry("no events", nil, false, false, false),
	)

	It("omits aggregates from json when not set", func() {
		report := NewImagesReport()

		data, err := report.ToJsonData()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring("Operations"))
		Expect(string(data)).NotTo(ContainSubstring("StageCache"))
		Expect(string(data)).NotTo(ContainSubstring("RegistryCache"))
		Expect(string(data)).NotTo(ContainSubstring("Recovery"))
	})
})

var _ = Describe("ImagesReport cache operations summary", func() {
	ctx := context.Background()

	It("serializes the cache counters next to the legacy cache event sections", func() {
		report := NewImagesReport()
		report.SetOperationsSummary(ctx, nil, []opstats.EventSummary{
			{Event: opstats.EventRegistryTagsCacheHit, Count: 3},
			{Event: opstats.EventStageBuilt, Count: 1},
		})
		report.SetCacheOperationsSummary(ctx, []opstats.CacheSummary{
			{Operation: opstats.OperationGitPatch, Layer: opstats.CacheLayerMemory, Hit: 3, Miss: 2, Bypass: 1, Shared: 1},
			{Operation: opstats.OperationGitPatch, Layer: opstats.CacheLayerDisk, Hit: 1, Miss: 1},
			{Operation: opstats.OperationDockerImageList, Layer: opstats.CacheLayerMemory, Bypass: 4},
		})

		data, err := report.ToJsonData()
		Expect(err).NotTo(HaveOccurred())

		decoded := decodeOperationsReport(data)
		// The layers of one operation are nested under the same operation key, so the
		// key stays the one Operations uses while each real layer keeps its own row.
		Expect(decoded.CacheOperations).To(Equal(map[string]map[string]ReportCacheOperationRecord{
			"git: patch": {
				"memory": {Lookups: 6, Hit: 3, Miss: 2, Bypass: 1, Shared: 1},
				"disk":   {Lookups: 2, Hit: 1, Miss: 1},
			},
			"docker: image list": {
				"memory": {Lookups: 4, Bypass: 4},
			},
		}))
		// The legacy sections keep their existing names and meanings.
		Expect(decoded.StageCache).To(Equal(map[string]int{"built": 1}))
		Expect(decoded.RegistryCache).To(Equal(map[string]int{"registry tags cache hit": 3}))
		// A redundant hit rate is derivable and not serialized.
		Expect(string(data)).NotTo(ContainSubstring("HitPercent"))
	})

	It("omits the cache section when no lookup was recorded", func() {
		report := NewImagesReport()
		report.SetCacheOperationsSummary(ctx, nil)

		data, err := report.ToJsonData()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring("CacheOperations"))
	})
})
