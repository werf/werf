package cleanup_test

import (
	"fmt"
	"os"
	"strings"

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

		It("should remove image by expiry days policy and skip deployed one (WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS)", func() {
			out := utils.SucceedCommandOutputString(
				testDirPath,
				"git",
				"rev-parse", "HEAD",
			)
			commit := strings.TrimSpace(out)

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"build-and-publish", "--tag-git-commit", commit,
			)

			tags := imagesRepoAllImageRepoTags("")
			Ω(len(tags)).Should(Equal(2))

			werfArgs := []string{"cleanup", "--git-commit-strategy-expiry-days", "0"}
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				werfArgs...,
			)

			tags = imagesRepoAllImageRepoTags("")
			Ω(len(tags)).Should(Equal(1))
			Ω(tags).Should(ContainElement(imagesRepo.ImageRepositoryTag("", "none")))
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
