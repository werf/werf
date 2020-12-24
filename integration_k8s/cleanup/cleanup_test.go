package cleanup_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/utils"
)

var _ = Describe("cleaning images and stages", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("default"), SuiteData.TestDirPath)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"init",
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"add", "-A",
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"commit", "-m", "Initial commit",
		)
	})

	Context("with deployed image", func() {
		BeforeEach(func() {
			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"build",
			)

			Ω(len(ImageMetadata(imageName))).Should(Equal(1))

			werfDeployArgs := []string{
				"converge",
				"--set", fmt.Sprintf("imageCredentials.registry=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY")),
				"--set", fmt.Sprintf("imageCredentials.username=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME")),
				"--set", fmt.Sprintf("imageCredentials.password=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD")),
			}
			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				werfDeployArgs...,
			)
		})

		When("KeepStageSetsBuiltWithinLastNHours policy is disabled", func() {
			BeforeEach(func() {
				SuiteData.Stubs.SetEnv("WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS", "0")
			})

			It("should not remove stages that are related with deployed image", func() {
				count := StagesCount()
				Ω(count).Should(Equal(2))

				utils.RunSucceedCommand(
					SuiteData.TestDirPath,
					SuiteData.WerfBinPath,
					"cleanup",
				)

				Ω(StagesCount()).Should(Equal(count))
			})
		})
	})
})
