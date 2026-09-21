package common

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/pkg/config"
)

var _ = Describe("build report operations option", func() {
	Describe("SetupBuildReportOperations", func() {
		It("defaults to false", func() {
			cmdData := &CmdData{}
			SetupBuildReportOperations(cmdData, &cobra.Command{})
			Expect(GetBuildReportOperations(cmdData)).To(BeFalse())
		})

		It("is enabled by the flag", func() {
			cmdData := &CmdData{}
			cmd := &cobra.Command{}
			SetupBuildReportOperations(cmdData, cmd)
			Expect(cmd.Flags().Parse([]string{"--build-report-operations"})).To(Succeed())
			Expect(GetBuildReportOperations(cmdData)).To(BeTrue())
		})

		It("is enabled by $WERF_BUILD_REPORT_OPERATIONS", func() {
			GinkgoT().Setenv("WERF_BUILD_REPORT_OPERATIONS", "1")
			cmdData := &CmdData{}
			SetupBuildReportOperations(cmdData, &cobra.Command{})
			Expect(GetBuildReportOperations(cmdData)).To(BeTrue())
		})
	})

	DescribeTable("wiring into build options",
		func(enabled bool) {
			cmdData := &CmdData{
				Dev:                   lo.ToPtr(false),
				BuildReportOperations: lo.ToPtr(enabled),
			}
			werfConfig := &config.WerfConfig{Meta: &config.Meta{}}

			buildOptions, err := GetBuildOptions(context.Background(), cmdData, werfConfig, config.ImagesToProcess{})
			Expect(err).NotTo(HaveOccurred())
			Expect(buildOptions.ReportOperations).To(Equal(enabled))

			builtOptions, err := GetShouldBeBuiltOptions(cmdData, werfConfig, config.ImagesToProcess{})
			Expect(err).NotTo(HaveOccurred())
			Expect(builtOptions.ReportOperations).To(Equal(enabled))
		},
		Entry("enabled", true),
		Entry("disabled", false),
	)
})
