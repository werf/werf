package host_cleaning

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/werf/werf/v2/pkg/host_cleaning/units"
)

var _ = Describe("HostCleanup", func() {
	DescribeTable("getRequirementInBytes",
		func(uv *units.UnitValue, defaultPercent, total, expected uint64) {
			Expect(getRequirementInBytes(uv, defaultPercent, total)).To(Equal(expected))
		},
		Entry("should use default value when UnitValue is nil",
			nil,
			uint64(70),
			uint64(1000),
			uint64(700),
		),
		Entry("should use UnitValue percentage: 50% of 1000",
			lo.Must(units.ParseUnitValue("50")),
			uint64(70),
			uint64(1000),
			uint64(500),
		),
		Entry("should use UnitValue absolute bytes: 300 of 1000",
			lo.Must(units.ParseUnitValue("300B")),
			uint64(70),
			uint64(1000),
			uint64(300),
		),
	)

	Context("HostCleanupOptions integration", func() {
		It("should correctly provide thresholds for valid percentage units", func() {
			opts := HostCleanupOptions{
				AllowedBackendStorageVolumeUsage:       lo.Must(units.ParseUnitValue("80")),
				AllowedBackendStorageVolumeUsageMargin: lo.Must(units.ParseUnitValue("10")),
			}

			total := uint64(1000)
			resUsage := getRequirementInBytes(opts.AllowedBackendStorageVolumeUsage, DefaultAllowedBackendStorageVolumeUsagePercentage, total)
			resMargin := getRequirementInBytes(opts.AllowedBackendStorageVolumeUsageMargin, DefaultAllowedBackendStorageVolumeUsageMarginPercentage, total)

			Expect(resUsage).To(Equal(uint64(800)))
			Expect(resMargin).To(Equal(uint64(100)))
		})

		It("should correctly provide thresholds for valid absolute units", func() {
			opts := HostCleanupOptions{
				AllowedBackendStorageVolumeUsage:       lo.Must(units.ParseUnitValue("500B")),
				AllowedBackendStorageVolumeUsageMargin: lo.Must(units.ParseUnitValue("50B")),
			}

			total := uint64(1000)
			resUsage := getRequirementInBytes(opts.AllowedBackendStorageVolumeUsage, DefaultAllowedBackendStorageVolumeUsagePercentage, total)
			resMargin := getRequirementInBytes(opts.AllowedBackendStorageVolumeUsageMargin, DefaultAllowedBackendStorageVolumeUsageMarginPercentage, total)

			Expect(resUsage).To(Equal(uint64(500)))
			Expect(resMargin).To(Equal(uint64(50)))
		})
	})
})
