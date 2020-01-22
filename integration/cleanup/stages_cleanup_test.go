package cleanup_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
)

var _ = Describe("cleaning stages", func() {
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

		stubs.SetEnv("WERF_IMAGES_REPO", registryProjectRepository)
		stubs.SetEnv("WERF_STAGES_STORAGE", ":local")

		stubs.SetEnv("WERF_WITHOUT_KUBE", "1")

		stubs.SetEnv("FROM_CACHE_VERSION", "x")
	})

	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	for _, werfArgs := range [][]string{
		{"stages", "cleanup"},
		{"cleanup"},
	} {
		commandToCheck := strings.Join(werfArgs, " ") + " command"
		commandWerfArgs := werfArgs

		Describe(commandToCheck, func() {
			It("should work properly with non-existent registry repository", func() {
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"stages", "cleanup",
				)
			})

			for _, disableStageCleanupDatePeriodPolicy := range []string{"0", "1"} {
				if disableStageCleanupDatePeriodPolicy == "1" {
					It("should not remove stages images related with built images in repository (WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY=1)", func() {
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build-and-publish", "--tag-git-commit", commit,
						)

						count := LocalProjectStagesCount()
						Ω(count).Should(Equal(count))

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							commandWerfArgs...,
						)

						Ω(LocalProjectStagesCount()).Should(Equal(count))
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

							stubs.SetEnv("WERF_LOG_PRETTY", "0")
						})

						It("should skip stage image with related running container", func() {
							out, err := utils.RunCommand(
								testDirPath,
								werfBinPath,
								commandWerfArgs...,
							)
							Ω(err).Should(Succeed())
							Ω(string(out)).Should(ContainSubstring("Skip image "))
							Ω(string(out)).Should(ContainSubstring("used by container"))
						})
					})
				}

				boundedPolicyValue := disableStageCleanupDatePeriodPolicy

				var itMsg string
				if disableStageCleanupDatePeriodPolicy == "0" {
					itMsg = fmt.Sprintf("should not remove unused stages images (WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY=0)")
				} else {
					itMsg = fmt.Sprintf("should remove unused stages images (WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY=1)")
				}

				It(itMsg, func() {
					stubs.SetEnv("WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY", boundedPolicyValue)

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build-and-publish", "--tag-git-commit", commit,
					)

					countAfterFirstBuild := LocalProjectStagesCount()
					Ω(countAfterFirstBuild).Should(Equal(countAfterFirstBuild))

					stubs.SetEnv("FROM_CACHE_VERSION", "fully rebuild")

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build-and-publish", "--tag-git-commit", commit,
					)

					countAfterSecondBuild := LocalProjectStagesCount()

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						commandWerfArgs...,
					)

					Ω(LocalProjectStagesCount()).Should(Equal(countAfterSecondBuild - countAfterFirstBuild))
				})
			}
		})
	}
})
