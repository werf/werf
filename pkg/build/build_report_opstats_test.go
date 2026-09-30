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
		Expect(decoded.StageCache).To(Equal(map[string]int{"built": 1, "found in repo stages storage": 2}))
		Expect(decoded.RegistryCache).To(Equal(map[string]int{"registry tags cache hit": 3, "registry tags shared result": 1}))
	})

	DescribeTable("omits a cache section that has no events",
		func(events []opstats.EventSummary, stageCachePresent, registryCachePresent bool) {
			report := NewImagesReport()
			report.SetOperationsSummary(ctx, nil, events)

			data, err := report.ToJsonData()
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.Contains(string(data), `"StageCache"`)).To(Equal(stageCachePresent))
			Expect(strings.Contains(string(data), `"RegistryCache"`)).To(Equal(registryCachePresent))
		},
		Entry("stage events only", []opstats.EventSummary{{Event: opstats.EventStageBuilt, Count: 1}}, true, false),
		Entry("registry events only", []opstats.EventSummary{{Event: opstats.EventRegistryTagsCacheHit, Count: 1}}, false, true),
		Entry("no events", nil, false, false),
	)

	It("omits aggregates from json when not set", func() {
		report := NewImagesReport()

		data, err := report.ToJsonData()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring("Operations"))
		Expect(string(data)).NotTo(ContainSubstring("StageCache"))
		Expect(string(data)).NotTo(ContainSubstring("RegistryCache"))
	})
})
