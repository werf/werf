package cleanup_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/utils"
	"github.com/werf/werf/pkg/docker_registry"
)

var _ = forEachDockerRegistryImplementation("cleanup command", func() {
	BeforeEach(func() {
		stubs.SetEnv("WERF_WITHOUT_KUBE", "1")
	})

	Describe("git history-based cleanup", func() {
		BeforeEach(func() {
			stubs.SetEnv("WERF_DISABLE_GITERMENISM", "1") // FIXME
			utils.CopyIn(utils.FixturePath("git_history_based_cleanup"), testDirPath)
			cleanupBeforeEachBase()
		})

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

		It("should keep several images that are related to one commit regardless of the keep policies (imagesPerReference.last=1)", func() {
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

		Context("keep policies", func() {
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

				It("should keep image metadata only for the latest commit (one branch)", func() {
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
			stubs.SetEnv("WERF_CONFIG", "werf_1.yaml")
			cleanupBeforeEachBase()
		})

		It("should work properly with non-existent/empty repo", func() {
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"cleanup",
			)
		})

		When("KeepStageSetsBuiltWithinLastNHours policy is disabled", func() {
			BeforeEach(func() {
				stubs.SetEnv("WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS", "0")
			})

			It("should remove unused stages", func() {
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

				stubs.SetEnv("WERF_CONFIG", "werf_2a.yaml") // full rebuild

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

			It("should not remove used stages", func() {
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

			When("there is running container based on werf stage", func() {
				BeforeEach(func() {
					if stagesStorage.Address() != ":local" {
						Skip(fmt.Sprintf("to test :local storage (%s)", stagesStorage.Address()))
					}

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

			Context("imports metadata", func() {
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

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"commit", "--allow-empty", "-m", "test",
					)

					stubs.SetEnv("WERF_CONFIG", "werf_2b.yaml") // full artifact rebuild
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
					if testImplementation != docker_registry.QuayImplementationName {
						Ω(countAfterSecondBuild).Should(Equal(6))
					}

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"cleanup",
					)

					if testImplementation != docker_registry.QuayImplementationName {
						countAfterCleanup := StagesCount()
						Ω(countAfterCleanup).Should(Equal(countAfterSecondBuild))
						Ω(len(ImportMetadataIDs())).Should(BeEquivalentTo(2))
					}
				})

				It("should remove unused artifact which does not have related import metadata", func() {
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

					if testImplementation != docker_registry.QuayImplementationName {
						countAfterCleanup := StagesCount()
						Ω(countAfterCleanup).Should(Equal(3))
						Ω(len(ImportMetadataIDs())).Should(BeEquivalentTo(0))
					}
				})
			})
		})

		When("KeepStageSetsBuiltWithinLastNHours policy is 2 hours (default)", func() {
			BeforeEach(func() {
				stubs.UnsetEnv("WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS")
			})

			It("should not remove unused stages that was built within 2 hours", func() {
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

				stubs.SetEnv("WERF_CONFIG", "werf_2a.yaml") // full rebuild

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				countAfterSecondBuild := StagesCount()
				if testImplementation != docker_registry.QuayImplementationName {
					Ω(countAfterSecondBuild).Should(Equal(8))
				}

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"cleanup",
				)

				if testImplementation != docker_registry.QuayImplementationName {
					Ω(StagesCount()).Should(Equal(countAfterSecondBuild))
				}
			})
		})
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
		"add", "werf*.yaml",
	)

	utils.RunSucceedCommand(
		testDirPath,
		"git",
		"commit", "-m", "Initial commit",
	)
}

func getHeadCommit() string {
	out := utils.SucceedCommandOutputString(
		testDirPath,
		"git",
		"rev-parse", "HEAD",
	)

	return strings.TrimSpace(out)
}

func gitHistoryBasedCleanupCheck(imageName string, expectedNumberOfMetadataTagsBefore, expectedNumberOfMetadataTagsAfter int, afterCleanupChecks ...func(map[string][]string)) {
	Ω(ImageMetadata(imageName)).Should(HaveLen(expectedNumberOfMetadataTagsBefore))

	utils.RunSucceedCommand(
		testDirPath,
		werfBinPath,
		"cleanup",
	)

	if testImplementation != docker_registry.QuayImplementationName {
		imageMetadata := ImageMetadata(imageName)
		Ω(imageMetadata).Should(HaveLen(expectedNumberOfMetadataTagsAfter))

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
