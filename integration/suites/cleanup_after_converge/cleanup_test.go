package cleanup_with_k8s_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

const (
	customTagValuePrefix = "user-custom-tag-"
	customTagValueFormat = "user-custom-tag-%v"
)

var _ = Describe("cleaning images and stages", func() {
	BeforeEach(func() {
		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("default"), "initial commit")
	})

	_ = AfterEach(func() {
		utils.RunSucceedCommand(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			"dismiss", "--with-namespace",
		)

		utils.RunSucceedCommand(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			"purge",
		)
	})

	When("KeepStageSetsBuiltWithinLastNHours policy is disabled", func() {
		BeforeEach(func() {
			SuiteData.Stubs.SetEnv("WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS", "0")
		})

		DescribeTable("should keep stages and custom tags", func(useCustomTag bool) {
			addCustomTagValue1 := fmt.Sprintf(customTagValueFormat, "1")
			addCustomTagValue2 := fmt.Sprintf(customTagValueFormat, "2")
			useCustomTagValue := addCustomTagValue1

			By("Do build")
			{
				utils.RunSucceedCommand(
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
					SuiteData.WerfBinPath,
					"build",
					"--add-custom-tag", addCustomTagValue1,
					"--add-custom-tag", addCustomTagValue2,
				)
			}

			By("Do deploy")
			{
				var werfArgs []string
				werfArgs = append(werfArgs, "converge")
				werfArgs = append(
					werfArgs,
					"--set", fmt.Sprintf("imageCredentials.registry=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY")),
					"--set", fmt.Sprintf("imageCredentials.username=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME")),
					"--set", fmt.Sprintf("imageCredentials.password=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD")),
				)
				werfArgs = append(werfArgs, "--require-built-images")

				if useCustomTag {
					werfArgs = append(werfArgs, "--use-custom-tag", useCustomTagValue)
				}

				utils.RunSucceedCommand(
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
					SuiteData.WerfBinPath,
					werfArgs...,
				)
			}

			By("Do cleanup and check result")
			{
				Ω(len(ImageMetadata(imageName))).Should(Equal(1))
				count := StagesCount()
				Ω(count).Should(Equal(2))
				Ω(len(CustomTags())).Should(Equal(2))
				Ω(len(CustomTagsMetadataList())).Should(Equal(2))

				utils.RunSucceedCommand(
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
					SuiteData.WerfBinPath,
					"cleanup",
				)

				Ω(StagesCount()).Should(Equal(count))
				Ω(len(CustomTags())).Should(Equal(2))
				Ω(len(CustomTagsMetadataList())).Should(Equal(2))
			}
		},
			Entry("deployed stage", false),
			Entry("deployed custom tag", true),
		)
	})
})
