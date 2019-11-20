// +build integration

package cleanup_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/integration/utils"
)

var _ = Describe("purging stages", func() {
	BeforeEach(func() {
		utils.CopyIn(fixturePath("default"), testDirPath)

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

		Ω(os.Setenv("WERF_IMAGES_REPO", registryProjectRepository)).Should(Succeed())
		Ω(os.Setenv("WERF_STAGES_STORAGE", ":local")).Should(Succeed())
	})

	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	for _, werfArgs := range [][]string{
		{"stages", "purge"},
		{"purge"},
	} {
		commandToCheck := strings.Join(werfArgs, " ") + " command"
		commandWerfArgs := werfArgs

		Describe(commandToCheck, func() {
			It("should remove project images", func() {
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				Ω(LocalProjectStagesCount()).Should(BeNumerically(">", 0))

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					commandWerfArgs...,
				)

				Ω(LocalProjectStagesCount()).Should(Equal(0))
			})

			Context("when there is running container based on werf image", func() {
				BeforeEach(func() {
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"run", "--docker-options", "-d", "--", "/bin/sleep", "30",
					)

					Ω(os.Setenv("WERF_LOG_PRETTY", "0")).Should(Succeed())
				})

				It("should fail with specific error", func() {
					out, err := utils.RunCommand(
						testDirPath,
						werfBinPath,
						commandWerfArgs...,
					)
					Ω(err).ShouldNot(Succeed())
					Ω(string(out)).Should(ContainSubstring("used by container"))
					Ω(string(out)).Should(ContainSubstring("Use --force option to remove all containers that are based on deleting werf docker images"))
				})

				It("should remove project images and container", func() {
					werfArgs := append(commandWerfArgs, "--force")
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						werfArgs...,
					)
				})
			})
		})
	}
})
