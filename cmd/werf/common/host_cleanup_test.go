package common

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func autoHostCleanupCmdData(args ...string) *CmdData {
	ginkgo.GinkgoT().Setenv("WERF_DISABLE_AUTO_HOST_CLEANUP", "false")

	command := &cobra.Command{}
	data := &CmdData{}
	SetupDisableAutoHostCleanup(data, command)
	SetupCheckBuiltImages(data, command)
	gomega.Expect(command.ParseFlags(args)).To(gomega.Succeed())
	return data
}

var _ = ginkgo.Describe("Automatic host cleanup decision", func() {
	ginkgo.DescribeTable("skips the cleanup when images are only checked", func(flag string) {
		var args []string
		if flag == "" {
			ginkgo.GinkgoT().Setenv("WERF_CHECK_BUILT_IMAGES", "true")
		} else {
			args = append(args, flag)
		}

		cmdData := autoHostCleanupCmdData(args...)

		gomega.Expect(GetCheckBuiltImages(cmdData)).To(gomega.BeTrue(), "check mode must be enabled by %q", flag)
		gomega.Expect(skipAutoHostCleanup(cmdData)).To(gomega.BeTrue())
	},
		ginkgo.Entry("check flag", "--check-built-images"),
		ginkgo.Entry("legacy flag", "--require-built-images"),
		ginkgo.Entry("short flag", "-Z"),
		ginkgo.Entry("environment", ""),
	)

	ginkgo.It("skips the cleanup when it is disabled explicitly", func() {
		gomega.Expect(skipAutoHostCleanup(autoHostCleanupCmdData("--disable-auto-host-cleanup"))).To(gomega.BeTrue())
	})

	ginkgo.It("runs the cleanup for an ordinary build", func() {
		gomega.Expect(skipAutoHostCleanup(autoHostCleanupCmdData())).To(gomega.BeFalse())
	})

	ginkgo.It("runs the cleanup for a command requiring built images", func() {
		ginkgo.GinkgoT().Setenv("WERF_DISABLE_AUTO_HOST_CLEANUP", "false")
		ginkgo.GinkgoT().Setenv("WERF_REQUIRE_BUILT_IMAGES", "true")

		command := &cobra.Command{}
		cmdData := &CmdData{}
		SetupDisableAutoHostCleanup(cmdData, command)
		SetupRequireBuiltImages(cmdData, command)
		gomega.Expect(command.ParseFlags(nil)).To(gomega.Succeed())

		gomega.Expect(GetRequireBuiltImages(cmdData)).To(gomega.BeTrue())
		gomega.Expect(skipAutoHostCleanup(cmdData)).To(gomega.BeFalse())
	})
})
