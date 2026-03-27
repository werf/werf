package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("host cleanup CLI wiring", func() {
	Describe("SetupAllowedBackendStorageVolumeUsageMargin", func() {
		It("keeps nil by default", func() {
			GinkgoT().Setenv("WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE_MARGIN", "")
			GinkgoT().Setenv("WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN", "")

			var cmdData CmdData
			cmd := &cobra.Command{Use: "test"}

			SetupAllowedBackendStorageVolumeUsageMargin(&cmdData, cmd)

			Expect(cmdData.AllowedBackendStorageVolumeUsageMargin).To(BeNil())
			Expect(cmdData.AllowedBackendStorageVolumeUsageMarginExplicit).NotTo(BeNil())
			Expect(*cmdData.AllowedBackendStorageVolumeUsageMarginExplicit).To(BeFalse())
		})

		It("marks env value as explicit", func() {
			GinkgoT().Setenv("WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE_MARGIN", "2GB")
			GinkgoT().Setenv("WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN", "")

			var cmdData CmdData
			cmd := &cobra.Command{Use: "test"}

			SetupAllowedBackendStorageVolumeUsageMargin(&cmdData, cmd)

			Expect(cmdData.AllowedBackendStorageVolumeUsageMargin).NotTo(BeNil())
			Expect(cmdData.AllowedBackendStorageVolumeUsageMarginExplicit).NotTo(BeNil())
			Expect(*cmdData.AllowedBackendStorageVolumeUsageMarginExplicit).To(BeTrue())
			Expect(cmdData.AllowedBackendStorageVolumeUsageMargin.FormatCLIValue()).To(Equal("2000000000B"))
		})

		It("marks flag value as explicit", func() {
			GinkgoT().Setenv("WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE_MARGIN", "")
			GinkgoT().Setenv("WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN", "")

			var cmdData CmdData
			cmd := &cobra.Command{Use: "test"}

			SetupAllowedBackendStorageVolumeUsageMargin(&cmdData, cmd)

			Expect(cmd.Flags().Parse([]string{"--allowed-backend-storage-volume-usage-margin=2GB"})).To(Succeed())
			Expect(cmdData.AllowedBackendStorageVolumeUsageMargin).NotTo(BeNil())
			Expect(cmdData.AllowedBackendStorageVolumeUsageMarginExplicit).NotTo(BeNil())
			Expect(*cmdData.AllowedBackendStorageVolumeUsageMarginExplicit).To(BeTrue())
			Expect(cmdData.AllowedBackendStorageVolumeUsageMargin.FormatCLIValue()).To(Equal("2000000000B"))
		})
	})

	Describe("SetupAllowedLocalCacheVolumeUsageMargin", func() {
		It("keeps nil by default", func() {
			GinkgoT().Setenv("WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN", "")

			var cmdData CmdData
			cmd := &cobra.Command{Use: "test"}

			SetupAllowedLocalCacheVolumeUsageMargin(&cmdData, cmd)

			Expect(cmdData.AllowedLocalCacheVolumeUsageMargin).To(BeNil())
			Expect(cmdData.AllowedLocalCacheVolumeUsageMarginExplicit).NotTo(BeNil())
			Expect(*cmdData.AllowedLocalCacheVolumeUsageMarginExplicit).To(BeFalse())
		})

		It("marks env value as explicit", func() {
			GinkgoT().Setenv("WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN", "2GB")

			var cmdData CmdData
			cmd := &cobra.Command{Use: "test"}

			SetupAllowedLocalCacheVolumeUsageMargin(&cmdData, cmd)

			Expect(cmdData.AllowedLocalCacheVolumeUsageMargin).NotTo(BeNil())
			Expect(cmdData.AllowedLocalCacheVolumeUsageMarginExplicit).NotTo(BeNil())
			Expect(*cmdData.AllowedLocalCacheVolumeUsageMarginExplicit).To(BeTrue())
			Expect(cmdData.AllowedLocalCacheVolumeUsageMargin.FormatCLIValue()).To(Equal("2000000000B"))
		})

		It("marks flag value as explicit", func() {
			GinkgoT().Setenv("WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN", "")

			var cmdData CmdData
			cmd := &cobra.Command{Use: "test"}

			SetupAllowedLocalCacheVolumeUsageMargin(&cmdData, cmd)

			Expect(cmd.Flags().Parse([]string{"--allowed-local-cache-volume-usage-margin=2GB"})).To(Succeed())
			Expect(cmdData.AllowedLocalCacheVolumeUsageMargin).NotTo(BeNil())
			Expect(cmdData.AllowedLocalCacheVolumeUsageMarginExplicit).NotTo(BeNil())
			Expect(*cmdData.AllowedLocalCacheVolumeUsageMarginExplicit).To(BeTrue())
			Expect(cmdData.AllowedLocalCacheVolumeUsageMargin.FormatCLIValue()).To(Equal("2000000000B"))
		})
	})
})
