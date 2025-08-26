package volumeutils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/volumeutils"
)

var _ = Describe("volume usage", func() {
	DescribeTable("should calculate bytes to free",
		func(vu volumeutils.VolumeUsage, targetVolumeUsagePercentage float64, expectedFreedBytes uint64) {
			Expect(vu.BytesToFree(targetVolumeUsagePercentage)).To(Equal(expectedFreedBytes))
		},
		Entry("when if vu.Percentage() - targetVolumeUsagePercentage = negative value",
			volumeutils.VolumeUsage{
				UsedBytes:  20,
				TotalBytes: 100,
			},
			50.00,
			uint64(0),
		),
	)
})
