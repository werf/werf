package build

import (
	"encoding/json"
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

		var decoded struct {
			Operations map[string]ReportOperationRecord
			StageCache map[string]int
		}
		Expect(json.Unmarshal(data, &decoded)).To(Succeed())

		Expect(decoded.Operations).To(Equal(map[string]ReportOperationRecord{
			"image push": {
				Count:            2,
				TotalTimeSeconds: 3,
				WallTimeSeconds:  2,
				AvgTimeSeconds:   1.5,
				MaxTimeSeconds:   2,
			},
		}))
		Expect(decoded.StageCache).To(Equal(map[string]int{"built": 1}))
	})

	It("omits aggregates from json when not set", func() {
		report := NewImagesReport()

		data, err := report.ToJsonData()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring("Operations"))
		Expect(string(data)).NotTo(ContainSubstring("StageCache"))
	})
})
