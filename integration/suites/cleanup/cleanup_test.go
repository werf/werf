package cleanup_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/test/pkg/suite_init"
	"github.com/werf/werf/test/pkg/utils"
)

const branchName = "test_branch"

var _ = Describe("cleanup command", func() {
	for _, iName := range suite_init.ContainerRegistryImplementationListToCheck(true) {
		implementationName := iName

		Context("["+implementationName+"]", func() {
			BeforeEach(perImplementationBeforeEach(implementationName))
			BeforeEach(func() {
				SuiteData.Stubs.SetEnv("WERF_WITHOUT_KUBE", "1")
			})

			Describe("git history-based cleanup", func() {
				BeforeEach(func() {
					SuiteData.Stubs.SetEnv("WERF_LOOSE_GITERMINISM", "1") // FIXME
					utils.CopyIn(utils.FixturePath("git_history_based_cleanup"), SuiteData.TestDirPath)
					cleanupBeforeEachBase()
				})

				It("should work only with remote references", func() {
					SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "1")

					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						"git",
						"checkout", "-b", "test",
					)

					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						"git",
						"push", "--set-upstream", "origin", "test",
					)

					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						SuiteData.WerfBinPath,
						"build",
					)

					gitHistoryBasedCleanupCheck(imageName, 1, 1)

					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						"git",
						"push", "origin", "--delete", "test",
					)

					gitHistoryBasedCleanupCheck(imageName, 1, 0)
				})

				It("should remove image from untracked branch", func() {
					SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "1")

					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						"git",
						"checkout", "-b", "some_branch",
					)

					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						"git",
						"push", "--set-upstream", "origin", "some_branch",
					)

					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						SuiteData.WerfBinPath,
						"build",
					)

					gitHistoryBasedCleanupCheck(imageName, 1, 0)
				})

				It("should keep several images that are related to one commit regardless of the keep policies (imagesPerReference.last=1)", func() {
					SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "9")

					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						"git",
						"checkout", "-b", "test",
					)

					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						"git",
						"push", "--set-upstream", "origin", "test",
					)

					SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")
					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						SuiteData.WerfBinPath,
						"build",
					)

					SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "2")
					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						SuiteData.WerfBinPath,
						"build",
					)

					gitHistoryBasedCleanupCheck(imageName, 2, 2)
				})

				Context("keep policies", func() {
					It("should remove image by imagesPerReference.last", func() {
						SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "2")

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"checkout", "-b", "test",
						)

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"push", "--set-upstream", "origin", "test",
						)

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"build",
						)

						gitHistoryBasedCleanupCheck(imageName, 1, 0)
					})

					It("should remove image by imagesPerReference.in", func() {
						SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "3")

						setLastCommitCommitterWhen(time.Now().Add(-(25 * time.Hour)))

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"checkout", "-b", "test",
						)

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"push", "--set-upstream", "origin", "test",
						)

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"build",
						)

						gitHistoryBasedCleanupCheck(imageName, 1, 0)

						setLastCommitCommitterWhen(time.Now().Add(-(23 * time.Hour)))

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"push", "--force",
						)

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"build",
						)

						gitHistoryBasedCleanupCheck(imageName, 1, 1)
					})

					It("should remove image by references.limit.in", func() {
						SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "4")

						setLastCommitCommitterWhen(time.Now().Add(-(13 * time.Hour)))

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"checkout", "-b", "test",
						)

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"push", "--set-upstream", "origin", "test",
						)

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"build",
						)

						gitHistoryBasedCleanupCheck(imageName, 1, 0)

						setLastCommitCommitterWhen(time.Now().Add(-(11 * time.Hour)))

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"push", "--set-upstream", "--force",
						)

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
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
								SuiteData.TestDirPath,
								"git",
								"checkout", "-b", ref1,
							)

							setLastCommitCommitterWhen(time.Now().Add(-(13 * time.Hour)))

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								"git",
								"push", "--set-upstream", "origin", ref1,
							)

							SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")
							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								SuiteData.WerfBinPath,
								"build",
							)

							stageID1 = resultStageID(imageName)
							_ = stageID1

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								"git",
								"checkout", "-b", ref2,
							)

							setLastCommitCommitterWhen(time.Now().Add(-(11 * time.Hour)))

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								"git",
								"push", "--set-upstream", "origin", ref2,
							)

							SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "2")
							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								SuiteData.WerfBinPath,
								"build",
							)

							stageID2 = resultStageID(imageName)

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								"git",
								"checkout", "-b", ref3,
							)

							setLastCommitCommitterWhen(time.Now())

							SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "3")
							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								SuiteData.WerfBinPath,
								"build",
							)

							stageID3 = resultStageID(imageName)

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								"git",
								"push", "--set-upstream", "origin", ref3,
							)
						})

						It("should remove image by references.limit.in OR references.limit.last", func() {
							SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "5")
							gitHistoryBasedCleanupCheck(
								imageName, 3, 2,
								func(imageMetadata map[string][]string) {
									Ω(imageMetadata).Should(HaveKey(stageID2))
									Ω(imageMetadata).Should(HaveKey(stageID3))
								},
							)
						})

						It("should remove image by references.limit.in AND references.limit.last", func() {
							SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "6")
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
								SuiteData.TestDirPath,
								"git",
								"checkout", "-b", "test",
							)

							setLastCommitCommitterWhen(time.Now().Add(-(13 * time.Hour)))

							SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")
							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								SuiteData.WerfBinPath,
								"build",
							)

							stageID1 = resultStageID(imageName)
							_ = stageID1

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								"git",
								"commit", "--allow-empty", "-m", "+",
							)

							setLastCommitCommitterWhen(time.Now().Add(-(11 * time.Hour)))

							SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "2")
							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								SuiteData.WerfBinPath,
								"build",
							)

							stageID2 = resultStageID(imageName)

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								"git",
								"commit", "--allow-empty", "-m", "+",
							)

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								"git",
								"push", "--set-upstream", "origin", "test",
							)

							SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "3")
							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								SuiteData.WerfBinPath,
								"build",
							)

							stageID3 = resultStageID(imageName)
						})

						It("should remove image by imagesPerReference.in OR imagesPerReference.last", func() {
							SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "7")
							gitHistoryBasedCleanupCheck(
								imageName, 3, 2,
								func(imageMetadata map[string][]string) {
									Ω(imageMetadata).Should(HaveKey(stageID2))
									Ω(imageMetadata).Should(HaveKey(stageID3))
								},
							)
						})

						It("should remove image by imagesPerReference.in AND imagesPerReference.last", func() {
							SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "8")
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
									SuiteData.TestDirPath,
									"git",
									"commit", "--allow-empty", "-m", "+",
								)

								utils.RunSucceedCommand(
									SuiteData.TestDirPath,
									SuiteData.WerfBinPath,
									"build",
								)
							}

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								"git",
								"push", "--set-upstream", "origin", branchName,
							)

							metaImagesCheckFunc(3, 1, func(commits []string) {
								Ω(commits).Should(ContainElement(utils.GetHeadCommit(SuiteData.TestDirPath)))
							})
						})

						It("should keep all image metadata (several branches)", func() {
							for i := 0; i < 3; i++ {
								utils.RunSucceedCommand(
									SuiteData.TestDirPath,
									"git",
									"commit", "--allow-empty", "-m", "+",
								)

								utils.RunSucceedCommand(
									SuiteData.TestDirPath,
									SuiteData.WerfBinPath,
									"build",
								)

								branch := fmt.Sprintf("test_%d", i)
								utils.RunSucceedCommand(
									SuiteData.TestDirPath,
									"git",
									"checkout", "-b", branch,
								)

								utils.RunSucceedCommand(
									SuiteData.TestDirPath,
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
					utils.CopyIn(utils.FixturePath("cleanup_unused_stages"), SuiteData.TestDirPath)
					SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_1.yaml")
					cleanupBeforeEachBase()
				})

				It("should work properly with non-existent/empty repo", func() {
					utils.RunSucceedCommand(
						SuiteData.TestDirPath,
						SuiteData.WerfBinPath,
						"cleanup",
					)
				})

				When("KeepStageSetsBuiltWithinLastNHours policy is disabled", func() {
					BeforeEach(func() {
						SuiteData.Stubs.SetEnv("WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS", "0")
					})

					It("should remove unused stages", func() {
						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"build",
						)

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"push", "--set-upstream", "origin", branchName,
						)

						countAfterFirstBuild := StagesCount()
						Ω(countAfterFirstBuild).Should(Equal(4))

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"commit", "--allow-empty", "-m", "test",
						)

						SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_2a.yaml") // full rebuild

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"build",
						)

						countAfterSecondBuild := StagesCount()

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"cleanup",
						)

						if SuiteData.TestImplementation != docker_registry.QuayImplementationName {
							Ω(StagesCount()).Should(Equal(countAfterSecondBuild - countAfterFirstBuild))
						}
					})

					It("should not remove used stages", func() {
						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"build",
						)

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"push", "--set-upstream", "origin", branchName,
						)

						count := StagesCount()
						Ω(count).Should(Equal(4))

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"cleanup",
						)

						Ω(StagesCount()).Should(Equal(count))
					})

					Context("custom tags", func() {
						It("should remove custom tag associated with the deleted stage", func() {
							customTag1 := "tag1"
							customTag2 := "tag2"

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								SuiteData.WerfBinPath,
								"build",
								"--add-custom-tag", fmt.Sprintf(customTagValueFormat, customTag1),
							)

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								"git",
								"push", "--set-upstream", "origin", branchName,
							)

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								"git",
								"commit", "--allow-empty", "-m", "test",
							)

							SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_2a.yaml") // full rebuild

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								SuiteData.WerfBinPath,
								"build",
								"--add-custom-tag", fmt.Sprintf(customTagValueFormat, customTag2),
							)

							customTags := CustomTags()
							Ω(customTags).Should(ContainElement(fmt.Sprintf(customTagValueFormat, customTag1)))
							Ω(customTags).Should(ContainElement(fmt.Sprintf(customTagValueFormat, customTag2)))
							Ω(len(CustomTagsMetadataList())).Should(Equal(2))

							utils.RunSucceedCommand(
								SuiteData.TestDirPath,
								SuiteData.WerfBinPath,
								"cleanup",
							)

							customTags = CustomTags()
							Ω(customTags).Should(ContainElement(fmt.Sprintf(customTagValueFormat, customTag1)))
							Ω(customTags).ShouldNot(ContainElement(fmt.Sprintf(customTagValueFormat, customTag2)))
							Ω(len(CustomTagsMetadataList())).Should(Equal(1))
						})
					})
				})

				When("KeepStageSetsBuiltWithinLastNHours policy is 2 hours (default)", func() {
					BeforeEach(func() {
						SuiteData.Stubs.UnsetEnv("WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS")
					})

					It("should not remove unused stages that was built within 2 hours", func() {
						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"build",
						)

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"push", "--set-upstream", "origin", branchName,
						)

						countAfterFirstBuild := StagesCount()
						Ω(countAfterFirstBuild).Should(Equal(4))

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							"git",
							"commit", "--allow-empty", "-m", "test",
						)

						SuiteData.Stubs.SetEnv("WERF_CONFIG", "werf_2a.yaml") // full rebuild

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"build",
						)

						countAfterSecondBuild := StagesCount()
						if SuiteData.TestImplementation != docker_registry.QuayImplementationName {
							Ω(countAfterSecondBuild).Should(Equal(8))
						}

						utils.RunSucceedCommand(
							SuiteData.TestDirPath,
							SuiteData.WerfBinPath,
							"cleanup",
						)

						if SuiteData.TestImplementation != docker_registry.QuayImplementationName {
							Ω(StagesCount()).Should(Equal(countAfterSecondBuild))
						}
					})
				})
			})
		})
	}
})

func resultStageID(imageName string) string {
	res := utils.SucceedCommandOutputString(
		SuiteData.TestDirPath,
		SuiteData.WerfBinPath,
		"stage", "image", imageName,
	)

	parts := strings.Split(strings.TrimSpace(res), ":")
	return parts[len(parts)-1]
}

func cleanupBeforeEachBase() {
	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"init", "--bare", "remote.git",
	)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"init",
	)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"remote", "add", "origin", "remote.git",
	)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"add", "werf*.yaml",
	)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"commit", "-m", "Initial commit",
	)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"checkout", "-b", branchName,
	)
}

func gitHistoryBasedCleanupCheck(imageName string, expectedNumberOfMetadataTagsBefore, expectedNumberOfMetadataTagsAfter int, afterCleanupChecks ...func(map[string][]string)) {
	Ω(ImageMetadata(imageName)).Should(HaveLen(expectedNumberOfMetadataTagsBefore))

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		SuiteData.WerfBinPath,
		"cleanup",
	)

	if SuiteData.TestImplementation != docker_registry.QuayImplementationName {
		imageMetadata := ImageMetadata(imageName)
		Ω(imageMetadata).Should(HaveLen(expectedNumberOfMetadataTagsAfter))

		for _, check := range afterCleanupChecks {
			check(imageMetadata)
		}
	}
}

func setLastCommitCommitterWhen(newDate time.Time) {
	_, _ = utils.RunCommandWithOptions(
		SuiteData.TestDirPath,
		"git",
		[]string{"commit", "--amend", "--allow-empty", "--no-edit", "--date", newDate.Format(time.RFC3339)},
		utils.RunCommandOptions{ShouldSucceed: true, ExtraEnv: []string{fmt.Sprintf("GIT_COMMITTER_DATE=%s", newDate.Format(time.RFC3339))}},
	)
}
