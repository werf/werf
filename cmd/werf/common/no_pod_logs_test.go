package common

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = ginkgo.Describe("--no-pod-logs flag", func() {
	ginkgo.It("defaults to false", func() {
		ginkgo.GinkgoT().Setenv("WERF_NO_POD_LOGS", "")

		cmdData := &CmdData{}
		cmd := &cobra.Command{}

		gomega.Expect(SetupTrackingFlags(cmdData, cmd)).To(gomega.Succeed())

		flag := cmd.PersistentFlags().Lookup("no-pod-logs")
		gomega.Expect(flag).NotTo(gomega.BeNil())
		gomega.Expect(flag.DefValue).To(gomega.Equal("false"))
		gomega.Expect(cmdData.NoPodLogs).To(gomega.BeFalse())
	})

	ginkgo.It("honors the $WERF_NO_POD_LOGS env var default", func() {
		ginkgo.GinkgoT().Setenv("WERF_NO_POD_LOGS", "true")

		cmdData := &CmdData{}
		cmd := &cobra.Command{}

		gomega.Expect(SetupTrackingFlags(cmdData, cmd)).To(gomega.Succeed())

		flag := cmd.PersistentFlags().Lookup("no-pod-logs")
		gomega.Expect(flag).NotTo(gomega.BeNil())
		gomega.Expect(flag.DefValue).To(gomega.Equal("true"))
		gomega.Expect(cmdData.NoPodLogs).To(gomega.BeTrue())
	})
})
