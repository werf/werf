package cleanup_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/testing/utils"
)

var _ = forEachDockerRegistryImplementation("cleaning stages", func() {
	var commit string

	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("stages_cleanup"), testDirPath)

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

		out := utils.SucceedCommandOutputString(
			testDirPath,
			"git",
			"rev-parse", "HEAD",
		)
		commit = strings.TrimSpace(out)

		stubs.SetEnv("WERF_WITHOUT_KUBE", "1")

		stubs.SetEnv("FROM_CACHE_VERSION", "x")
	})

	for _, werfCommand := range [][]string{
		{"stages", "cleanup"},
		{"cleanup"},
	} {
		description := strings.Join(werfCommand, " ") + " command"
		werfCommand := werfCommand

		Describe(description, func() {
			It("should work properly with non-existent/empty stages storage", func() {
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"stages", "cleanup",
				)
			})

			for _, disableStageCleanupDatePeriodPolicy := range []string{"0", "1"} {
				BeforeEach(func() {
					value := disableStageCleanupDatePeriodPolicy
					stubs.SetEnv("WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY", value)
				})

				if disableStageCleanupDatePeriodPolicy == "1" {
					It("should not remove stages that are related with images in images repo (WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY=1)", func() {
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build-and-publish", "--tag-git-commit", commit,
						)

						count := stagesStorageRepoImagesCount()
						Ω(count).Should(Equal(4))

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfCommand...,
						)

						Ω(stagesStorageRepoImagesCount()).Should(Equal(count))
					})

					Context("when there is running container that is based on werf image", func() {
						BeforeEach(func() {
							if stagesStorage.Address() != ":local" {
								Skip(fmt.Sprintf("to test :local stages storage (%s)", stagesStorage.Address()))
							}
						})

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

							stubs.SetEnv("WERF_LOG_PRETTY", "0")
						})

						It("should skip stage with related running container", func() {
							out, err := utils.RunCommand(
								testDirPath,
								werfBinPath,
								werfCommand...,
							)
							Ω(err).Should(Succeed())
							Ω(string(out)).Should(ContainSubstring("Skip image "))
							Ω(string(out)).Should(ContainSubstring("used by container"))
						})
					})
				}

				var itMsg string
				if disableStageCleanupDatePeriodPolicy == "0" {
					itMsg = fmt.Sprintf("should not remove unused stages (WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY=0)")
				} else {
					itMsg = fmt.Sprintf("should remove unused stages (WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY=1)")
				}

				It(itMsg, func() {
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build-and-publish", "--tag-git-commit", commit,
					)

					countAfterFirstBuild := stagesStorageRepoImagesCount()
					Ω(countAfterFirstBuild).Should(Equal(4))

					stubs.SetEnv("FROM_CACHE_VERSION", "full rebuild")

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build-and-publish", "--tag-git-commit", commit,
					)

					countAfterSecondBuild := stagesStorageRepoImagesCount()

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						werfCommand...,
					)

					Ω(stagesStorageRepoImagesCount()).Should(Equal(countAfterSecondBuild - countAfterFirstBuild))
				})
			}
		})
	}
})
