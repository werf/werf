package build

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/opstats"
)

var _ = Describe("ImagesReport operations summary", func() {
	It("serializes operations and stage cache aggregates to json", func() {
		report := NewImagesReport()
		report.SetOperationsSummary(
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
				{Event: opstats.EventStageBuilt, Count: 1},
			},
		)

		data, err := report.ToJsonData()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring(`"image push"`))
		Expect(string(data)).To(ContainSubstring(`"TotalTimeSeconds": 3`))
		Expect(string(data)).To(ContainSubstring(`"built": 1`))
	})

	It("omits aggregates from json when not set", func() {
		report := NewImagesReport()

		data, err := report.ToJsonData()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring("Operations"))
		Expect(string(data)).NotTo(ContainSubstring("StageCache"))
	})
})
