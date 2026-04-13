package host_cleaning

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/cmd/werf/common/units"
)

var _ = Describe("HostCleanup", func() {
	DescribeTable("getRequirementInBytes",
		func(sv *units.UnitValue, defaultPercent, totalBytes, expected uint64) {
			actual := getRequirementInBytes(sv, defaultPercent, totalBytes)
			Expect(actual).To(Equal(expected))
		},
		Entry("should use UnitValue percentage: 50% of 1000",
			&units.UnitValue{Value: 50, IsBytes: false},
			uint64(70),
			uint64(1000),
			uint64(500),
		),
		Entry("should use UnitValue absolute bytes: 300 of 1000",
			&units.UnitValue{Value: 300, IsBytes: true},
			uint64(70),
			uint64(1000),
			uint64(300),
		),
		Entry("should use default percentage when nil: 70% of 1000",
			nil,
			uint64(70),
			uint64(1000),
			uint64(700),
		),
	)

	Context("HostCleanupOptions integration", func() {
		It("should correctly provide thresholds for valid percentage units", func() {
			opts := HostCleanupOptions{
				AllowedBackendStorageVolumeUsage:       &units.UnitValue{Value: 80, IsBytes: false},
				AllowedBackendStorageVolumeUsageMargin: &units.UnitValue{Value: 10, IsBytes: false},
			}

			total := uint64(1000)
			usage := getRequirementInBytes(opts.AllowedBackendStorageVolumeUsage, DefaultAllowedBackendStorageVolumeUsagePercentage, total)
			margin := getRequirementInBytes(opts.AllowedBackendStorageVolumeUsageMargin, DefaultAllowedBackendStorageVolumeUsageMarginPercentage, total)

			Expect(usage).To(Equal(uint64(800)))
			Expect(margin).To(Equal(uint64(100)))
		})

		It("should correctly provide thresholds for valid absolute units", func() {
			opts := HostCleanupOptions{
				AllowedBackendStorageVolumeUsage:       &units.UnitValue{Value: 500, IsBytes: true},
				AllowedBackendStorageVolumeUsageMargin: &units.UnitValue{Value: 50, IsBytes: true},
			}

			total := uint64(1000)
			usage := getRequirementInBytes(opts.AllowedBackendStorageVolumeUsage, DefaultAllowedBackendStorageVolumeUsagePercentage, total)
			margin := getRequirementInBytes(opts.AllowedBackendStorageVolumeUsageMargin, DefaultAllowedBackendStorageVolumeUsageMarginPercentage, total)

			Expect(usage).To(Equal(uint64(500)))
			Expect(margin).To(Equal(uint64(50)))
		})
	})
})
