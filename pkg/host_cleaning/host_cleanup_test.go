package host_cleaning

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	thresholdpkg "github.com/werf/werf/v2/pkg/cleaning/threshold"
)

var _ = Describe("host cleanup threshold resolving", func() {
	Describe("resolveBackendStorageVolumeUsageThresholds", func() {
		It("uses default percentage margin", func() {
			threshold, margin, err := resolveBackendStorageVolumeUsageThresholds(loToPtr(thresholdpkg.NewPercentage(70)), nil, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(threshold).To(Equal(thresholdpkg.NewPercentage(70)))
			Expect(margin).To(Equal(DefaultAllowedBackendStorageVolumeUsageMarginThreshold()))
		})

		It("uses zero bytes margin for implicit bytes thresholds", func() {
			threshold, margin, err := resolveBackendStorageVolumeUsageThresholds(loToPtr(thresholdpkg.NewBytes(10_000_000_000)), nil, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(threshold).To(Equal(thresholdpkg.NewBytes(10_000_000_000)))
			Expect(margin).To(Equal(thresholdpkg.NewBytes(0)))
		})

		It("returns an error for explicitly mixed formats", func() {
			_, _, err := resolveBackendStorageVolumeUsageThresholds(
				loToPtr(thresholdpkg.NewBytes(10_000_000_000)),
				loToPtr(thresholdpkg.NewPercentage(5)),
				true,
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must use the same format"))
		})

		It("uses threshold-type default when mixed margin was not explicit", func() {
			threshold, margin, err := resolveBackendStorageVolumeUsageThresholds(
				loToPtr(thresholdpkg.NewBytes(10_000_000_000)),
				loToPtr(thresholdpkg.NewPercentage(5)),
				false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(threshold).To(Equal(thresholdpkg.NewBytes(10_000_000_000)))
			Expect(margin).To(Equal(thresholdpkg.NewBytes(0)))
		})

		It("passes explicit same-format values", func() {
			threshold, margin, err := resolveBackendStorageVolumeUsageThresholds(
				loToPtr(thresholdpkg.NewBytes(10_000_000_000)),
				loToPtr(thresholdpkg.NewBytes(2_000_000_000)),
				true,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(threshold).To(Equal(thresholdpkg.NewBytes(10_000_000_000)))
			Expect(margin).To(Equal(thresholdpkg.NewBytes(2_000_000_000)))
		})
	})

	Describe("resolveLocalCacheVolumeUsageThresholds", func() {
		It("uses default percentage margin", func() {
			threshold, margin, err := resolveLocalCacheVolumeUsageThresholds(loToPtr(thresholdpkg.NewPercentage(70)), nil, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(threshold).To(Equal(thresholdpkg.NewPercentage(70)))
			Expect(margin).To(Equal(DefaultAllowedLocalCacheVolumeUsageMarginThreshold()))
		})

		It("uses zero bytes margin for implicit bytes thresholds", func() {
			threshold, margin, err := resolveLocalCacheVolumeUsageThresholds(loToPtr(thresholdpkg.NewBytes(10_000_000_000)), nil, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(threshold).To(Equal(thresholdpkg.NewBytes(10_000_000_000)))
			Expect(margin).To(Equal(thresholdpkg.NewBytes(0)))
		})

		It("returns an error for explicitly mixed formats", func() {
			_, _, err := resolveLocalCacheVolumeUsageThresholds(
				loToPtr(thresholdpkg.NewBytes(10_000_000_000)),
				loToPtr(thresholdpkg.NewPercentage(5)),
				true,
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("--allowed-local-cache-volume-usage"))
		})
	})
})

func loToPtr[T any](v T) *T {
	return &v
}
