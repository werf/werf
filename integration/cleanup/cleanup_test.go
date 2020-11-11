package cleanup_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/testing/utils"
)

var _ = forEachDockerRegistryImplementation("cleanup command", func() {
	BeforeEach(func() {
		stubs.SetEnv("WERF_WITHOUT_KUBE", "1")
	})

	Describe("git history-based cleanup", func() {
		BeforeEach(func() {
			utils.CopyIn(utils.FixturePath("git_history_based_cleanup"), testDirPath)
			cleanupBeforeEachBase()
		})

		Context("stage IDs cleanup", func() {
			It("should work only with remote references", func() {
				stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "1")

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"checkout", "-b", "test",
				)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"push", "--set-upstream", "origin", "test",
				)

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				gitHistoryBasedCleanupCheck(imageName, 1, 1)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"push", "origin", "--delete", "test",
				)

				gitHistoryBasedCleanupCheck(imageName, 1, 0)
			})

			It("should remove image from untracked branch", func() {
				stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "1")

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"checkout", "-b", "some_branch",
				)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"push", "--set-upstream", "origin", "some_branch",
				)

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				gitHistoryBasedCleanupCheck(imageName, 1, 0)
			})

			It("should remove image by imagesPerReference.last", func() {
				stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "2")

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"checkout", "-b", "test",
				)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"push", "--set-upstream", "origin", "test",
				)

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				gitHistoryBasedCleanupCheck(imageName, 1, 0)
			})

			It("should remove image by imagesPerReference.in", func() {
				stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "3")

				setLastCommitCommitterWhen(time.Now().Add(-(25 * time.Hour)))

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"checkout", "-b", "test",
				)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"push", "--set-upstream", "origin", "test",
				)

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				gitHistoryBasedCleanupCheck(imageName, 1, 0)

				setLastCommitCommitterWhen(time.Now().Add(-(23 * time.Hour)))

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"push", "--force",
				)

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				gitHistoryBasedCleanupCheck(imageName, 1, 1)
			})

			It("should remove image by references.limit.in", func() {
				stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "4")

				setLastCommitCommitterWhen(time.Now().Add(-(13 * time.Hour)))

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"checkout", "-b", "test",
				)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"push", "--set-upstream", "origin", "test",
				)

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				gitHistoryBasedCleanupCheck(imageName, 1, 0)

				setLastCommitCommitterWhen(time.Now().Add(-(11 * time.Hour)))

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"push", "--set-upstream", "--force",
				)

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				gitHistoryBasedCleanupCheck(imageName, 1, 1)
			})

			Context("references.limit.*", func() {
				const (
					ref1 = "test1"
					ref2 = "test2"
					ref3 = "test3"
				)

				var (
					stageID1 string
					stageID2 string
					stageID3 string
				)

				BeforeEach(func() {
					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"checkout", "-b", ref1,
					)

					setLastCommitCommitterWhen(time.Now().Add(-(13 * time.Hour)))

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"push", "--set-upstream", "origin", ref1,
					)

					stubs.SetEnv("FROM_CACHE_VERSION", "1")
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)

					stageID1 = resultStageID(imageName)
					_ = stageID1

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"checkout", "-b", ref2,
					)

					setLastCommitCommitterWhen(time.Now().Add(-(11 * time.Hour)))

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"push", "--set-upstream", "origin", ref2,
					)

					stubs.SetEnv("FROM_CACHE_VERSION", "2")
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)

					stageID2 = resultStageID(imageName)

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"checkout", "-b", ref3,
					)

					setLastCommitCommitterWhen(time.Now())

					stubs.SetEnv("FROM_CACHE_VERSION", "3")
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)

					stageID3 = resultStageID(imageName)

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"push", "--set-upstream", "origin", ref3,
					)
				})

				It("should remove image by references.limit.in OR references.limit.last", func() {
					stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "5")
					gitHistoryBasedCleanupCheck(
						imageName, 3, 2,
						func(imageMetadata map[string][]string) {
							Ω(imageMetadata).Should(HaveKey(stageID2))
							Ω(imageMetadata).Should(HaveKey(stageID3))
						},
					)
				})

				It("should remove image by references.limit.in AND references.limit.last", func() {
					stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "6")
					gitHistoryBasedCleanupCheck(
						imageName, 3, 1,
						func(imageMetadata map[string][]string) {
							Ω(imageMetadata).Should(HaveKey(stageID3))
						},
					)
				})
			})

			Context("imagesPerReference.*", func() {
				var (
					stageID1 string
					stageID2 string
					stageID3 string
				)

				BeforeEach(func() {
					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"checkout", "-b", "test",
					)

					setLastCommitCommitterWhen(time.Now().Add(-(13 * time.Hour)))

					stubs.SetEnv("FROM_CACHE_VERSION", "1")
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)

					stageID1 = resultStageID(imageName)
					_ = stageID1

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"commit", "--allow-empty", "-m", "+",
					)

					setLastCommitCommitterWhen(time.Now().Add(-(11 * time.Hour)))

					stubs.SetEnv("FROM_CACHE_VERSION", "2")
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)

					stageID2 = resultStageID(imageName)

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"commit", "--allow-empty", "-m", "+",
					)

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"push", "--set-upstream", "origin", "test",
					)

					stubs.SetEnv("FROM_CACHE_VERSION", "3")
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)

					stageID3 = resultStageID(imageName)
				})

				It("should remove image by imagesPerReference.in OR imagesPerReference.last", func() {
					stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "7")
					gitHistoryBasedCleanupCheck(
						imageName, 3, 2,
						func(imageMetadata map[string][]string) {
							Ω(imageMetadata).Should(HaveKey(stageID2))
							Ω(imageMetadata).Should(HaveKey(stageID3))
						},
					)
				})

				It("should remove image by imagesPerReference.in AND imagesPerReference.last", func() {
					stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "8")
					gitHistoryBasedCleanupCheck(
						imageName, 3, 1,
						func(imageMetadata map[string][]string) {
							Ω(imageMetadata).Should(HaveKey(stageID3))
						},
					)
				})
			})

			XIt("should keep all stages that are related to one commit regardless of the keep policies", func() {
				stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "9")

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"checkout", "-b", "test",
				)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"push", "--set-upstream", "origin", "test",
				)

				stubs.SetEnv("FROM_CACHE_VERSION", "1")
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				stubs.SetEnv("FROM_CACHE_VERSION", "2")
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				gitHistoryBasedCleanupCheck(imageName, 2, 2)
			})
		})

		Context("images metadata cleanup", func() {
			When("one content digest", func() {
				metaImagesCheckFunc := func(before, after int, afterExtraChecks ...func(commits []string)) {
					imageMetadata := ImageMetadata(imageName)
					for _, commitList := range imageMetadata {
						Ω(commitList).Should(HaveLen(before))
					}

					gitHistoryBasedCleanupCheck(
						imageName, 1, 1,
						func(imageMetadata map[string][]string) {
							for _, commitList := range imageMetadata {
								Ω(commitList).Should(HaveLen(after))

								for _, check := range afterExtraChecks {
									check(commitList)
								}
							}
						},
					)
				}

				It("should remove all image metadata except the latest (one branch)", func() {
					for i := 0; i < 3; i++ {
						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"commit", "--allow-empty", "-m", "+",
						)

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build",
						)
					}

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"push", "--set-upstream", "origin", "master",
					)

					metaImagesCheckFunc(3, 1, func(commits []string) {
						Ω(commits).Should(ContainElement(getHeadCommit()))
					})
				})

				It("should keep all image metadata (several branches)", func() {
					for i := 0; i < 3; i++ {
						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"commit", "--allow-empty", "-m", "+",
						)

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build",
						)

						branch := fmt.Sprintf("test_%d", i)
						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"checkout", "-b", branch,
						)

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"push", "--set-upstream", "origin", branch,
						)
					}

					metaImagesCheckFunc(3, 3)
				})
			})
		})
	})

	Describe("cleanup unused stages", func() {
		BeforeEach(func() {
			utils.CopyIn(utils.FixturePath("cleanup_unused_stages"), testDirPath)

			cleanupBeforeEachBase()

			stubs.SetEnv("FROM_CACHE_VERSION", "x")
		})

		It("should work properly with non-existent/empty stages storage", func() {
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"cleanup",
			)
		})

		for _, disableStageCleanupDatePeriodPolicy := range []string{"0", "1"} {
			BeforeEach(func() {
				value := disableStageCleanupDatePeriodPolicy
				stubs.SetEnv("WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY", value)
			})

			if disableStageCleanupDatePeriodPolicy == "1" {
				It("should not remove stages that are related with image metadata (WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY=1)", func() {
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"push", "--set-upstream", "origin", "master",
					)

					count := StagesCount()
					Ω(count).Should(Equal(4))

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"cleanup",
					)

					Ω(StagesCount()).Should(Equal(count))
				})

				Context("when there is running container that is based on werf image", func() {
					BeforeEach(func() {
						if stagesStorage.Address() != ":local" {
							Skip(fmt.Sprintf("to test :local storage (%s)", stagesStorage.Address()))
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
					})

					It("should skip stage with related running container", func() {
						stubs.SetEnv("WERF_LOG_PRETTY", "0")

						out, err := utils.RunCommand(
							testDirPath,
							werfBinPath,
							"cleanup",
						)
						Ω(err).Should(Succeed())
						Ω(string(out)).Should(ContainSubstring("Skip image "))
						Ω(string(out)).Should(ContainSubstring("used by container"))
					})
				})

				Context("imports metadata (WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY=1)", func() {
					It("should keep used artifact", func() {
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build",
						)

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"push", "--set-upstream", "origin", "master",
						)

						countAfterFirstBuild := StagesCount()
						Ω(countAfterFirstBuild).Should(Equal(4))

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"cleanup",
						)

						countAfterCleanup := StagesCount()
						Ω(countAfterCleanup).Should(Equal(4))
						Ω(len(ImportMetadataIDs())).Should(BeEquivalentTo(1))
					})

					It("should keep both artifacts by identical import checksum", func() {
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build",
						)

						countAfterFirstBuild := StagesCount()
						Ω(countAfterFirstBuild).Should(Equal(4))

						stubs.SetEnv("ARTIFACT_FROM_CACHE_VERSION", "full rebuild")
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build",
						)

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"push", "--set-upstream", "origin", "master",
						)

						countAfterSecondBuild := StagesCount()
						Ω(countAfterSecondBuild).Should(Equal(6))

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"cleanup",
						)

						countAfterCleanup := StagesCount()
						Ω(countAfterCleanup).Should(Equal(countAfterSecondBuild))
						Ω(len(ImportMetadataIDs())).Should(BeEquivalentTo(2))
					})

					It("should remove unused artifact (without related import metadata)", func() {
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build",
						)

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"push", "--set-upstream", "origin", "master",
						)

						countAfterFirstBuild := StagesCount()
						Ω(countAfterFirstBuild).Should(Equal(4))
						Ω(len(ImportMetadataIDs())).Should(BeEquivalentTo(1))

						RmImportMetadata(ImportMetadataIDs()[0])

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"cleanup",
						)

						countAfterCleanup := StagesCount()
						Ω(countAfterCleanup).Should(Equal(3))
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
					"build",
				)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"push", "--set-upstream", "origin", "master",
				)

				countAfterFirstBuild := StagesCount()
				Ω(countAfterFirstBuild).Should(Equal(4))

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"commit", "--allow-empty", "-m", "test",
				)

				stubs.SetEnv("FROM_CACHE_VERSION", "full rebuild")

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				countAfterSecondBuild := StagesCount()

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"cleanup",
				)

				if testImplementation != docker_registry.QuayImplementationName {
					Ω(StagesCount()).Should(Equal(countAfterSecondBuild - countAfterFirstBuild))
				}
			})
		}
	})
})

func resultStageID(imageName string) string {
	res := utils.SucceedCommandOutputString(
		testDirPath,
		werfBinPath,
		"stage", "image", imageName,
	)

	parts := strings.Split(strings.TrimSpace(res), ":")
	return parts[len(parts)-1]
}

func cleanupBeforeEachBase() {
	utils.RunSucceedCommand(
		testDirPath,
		"git",
		"init", "--bare", "remote.git",
	)

	utils.RunSucceedCommand(
		testDirPath,
		"git",
		"init",
	)

	utils.RunSucceedCommand(
		testDirPath,
		"git",
		"remote", "add", "origin", "remote.git",
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
}

func getHeadCommit() string {
	out := utils.SucceedCommandOutputString(
		testDirPath,
		"git",
		"rev-parse", "HEAD",
	)

	return strings.TrimSpace(out)
}

func gitHistoryBasedCleanupCheck(imageName string, expectedNumberOfTagsBefore, expectedNumberOfTagsAfter int, afterCleanupChecks ...func(map[string][]string)) {
	Ω(ImageMetadata(imageName)).Should(HaveLen(expectedNumberOfTagsBefore))

	utils.RunSucceedCommand(
		testDirPath,
		werfBinPath,
		"cleanup",
	)

	if testImplementation != docker_registry.QuayImplementationName {
		imageMetadata := ImageMetadata(imageName)
		Ω(imageMetadata).Should(HaveLen(expectedNumberOfTagsAfter))

		if afterCleanupChecks != nil {
			for _, check := range afterCleanupChecks {
				check(imageMetadata)
			}
		}
	}
}

func setLastCommitCommitterWhen(newDate time.Time) {
	_, _ = utils.RunCommandWithOptions(
		testDirPath,
		"git",
		[]string{"commit", "--amend", "--allow-empty", "--no-edit", "--date", newDate.Format(time.RFC3339)},
		utils.RunCommandOptions{ShouldSucceed: true, ExtraEnv: []string{fmt.Sprintf("GIT_COMMITTER_DATE=%s", newDate.Format(time.RFC3339))}},
	)
}
