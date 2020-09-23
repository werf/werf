package cleanup_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/testing/utils"
)

var _ = Describe("cleaning images and stages", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("default"), testDirPath)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"init",
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"add", "werf.yaml",
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"commit", "-m", "Initial commit",
		)

		stubs.SetEnv("WERF_SKIP_GIT_FETCH", "1")
		stubs.SetEnv("WERF_GIT_HISTORY_BASED_CLEANUP", "0")
	})

	Context("with deployed image", func() {
		BeforeEach(func() {
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"build-and-publish", "--tag-git-commit", "none",
			)

			tags := imagesRepoAllImageRepoTags("")
			Ω(len(tags)).Should(Equal(1))

			werfDeployArgs := []string{
				"deploy",
				"--tag-git-commit", "none",
				"--env", "test",
				"--set", fmt.Sprintf("imageCredentials.registry=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY")),
				"--set", fmt.Sprintf("imageCredentials.username=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME")),
				"--set", fmt.Sprintf("imageCredentials.password=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD")),
			}
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				werfDeployArgs...,
			)
		})

		It("should not remove stages that are related with deployed image (WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY=1)", func() {
			stubs.SetEnv("WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY", "1")

			count := stagesStorageRepoImagesCount()
			Ω(count).Should(Equal(2))

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"cleanup",
			)

			Ω(stagesStorageRepoImagesCount()).Should(Equal(count))
		})
	})
})
