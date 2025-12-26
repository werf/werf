package cleanup_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
)

const branchName = "test_branch"

var _ = Describe("cleanup command", func() {
	for _, iName := range suite_init.ContainerRegistryImplementationListToCheck(true) {
		implementationName := iName

		Context("["+implementationName+"]", func() {
			BeforeEach(perImplementationBeforeEach(implementationName))

			Describe("default", func() {
				BeforeEach(func(ctx SpecContext) {
					SuiteData.Stubs.SetEnv("WERF_LOOSE_GITERMINISM", "1") // FIXME
					utils.CopyIn(utils.FixturePath("cleanup"), SuiteData.TestDirPath)
					cleanupBeforeEachBase(ctx)
				})

				It("should work properly with non-existent/empty repo", func(ctx SpecContext) {
					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "cleanup")
				})

				It("should not remove unused stages that was built within 2 hours", func(ctx SpecContext) {
					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", branchName)

					countAfterFirstBuild := StagesCount(ctx)
					Expect(countAfterFirstBuild).Should(Equal(4))

					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "--allow-empty", "-m", "test")

					SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "REBUILD")

					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

					countAfterSecondBuild := StagesCount(ctx)
					if SuiteData.TestImplementation != docker_registry.QuayImplementationName {
						Expect(countAfterSecondBuild).Should(Equal(8))
					}

					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "cleanup")

					if SuiteData.TestImplementation != docker_registry.QuayImplementationName {
						Expect(StagesCount(ctx)).Should(Equal(countAfterSecondBuild))
					}
				})
			})

			Describe("git history based policy", func() {
				BeforeEach(func(ctx SpecContext) {
					SuiteData.Stubs.SetEnv("WERF_LOOSE_GITERMINISM", "1") // FIXME
					utils.CopyIn(utils.FixturePath("git_history_based_policy"), SuiteData.TestDirPath)
					cleanupBeforeEachBase(ctx)
				})

				It("should remove unused stages", func(ctx SpecContext) {
					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", branchName)

					countAfterFirstBuild := StagesCount(ctx)
					Expect(countAfterFirstBuild).Should(Equal(4))

					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "--allow-empty", "-m", "test")

					SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "REBUILD")

					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

					countAfterSecondBuild := StagesCount(ctx)

					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "cleanup")

					if SuiteData.TestImplementation != docker_registry.QuayImplementationName {
						Expect(StagesCount(ctx)).Should(Equal(countAfterSecondBuild - countAfterFirstBuild))
					}
				})

				It("should not remove used stages", func(ctx SpecContext) {
					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", branchName)

					count := StagesCount(ctx)
					Expect(count).Should(Equal(4))

					utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "cleanup")

					Expect(StagesCount(ctx)).Should(Equal(count))
				})

				Context("image metadata", func() {
					It("should work only with remote references", func(ctx SpecContext) {
						SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "1")

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", "test")

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", "test")

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

						gitHistoryBasedCleanupCheck(ctx, imageName, 1, 1)

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "origin", "--delete", "test")

						gitHistoryBasedCleanupCheck(ctx, imageName, 1, 0)
					})

					It("should remove image from untracked branch", func(ctx SpecContext) {
						SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "1")

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", "some_branch")

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", "some_branch")

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

						gitHistoryBasedCleanupCheck(ctx, imageName, 1, 0)
					})

					It("should keep several images that are related to one commit regardless of the keep policies (imagesPerReference.last=1)", func(ctx SpecContext) {
						SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "9")

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", "test")

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", "test")

						SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")
						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

						SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "2")
						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

						gitHistoryBasedCleanupCheck(ctx, imageName, 2, 2)
					})

					Context("keep policies", func() {
						It("should remove image by imagesPerReference.last", func(ctx SpecContext) {
							SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "2")

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", "test")

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", "test")

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

							gitHistoryBasedCleanupCheck(ctx, imageName, 1, 0)
						})

						It("should remove image by imagesPerReference.in", func(ctx SpecContext) {
							SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "3")

							setLastCommitCommitterWhen(ctx, time.Now().Add(-(25 * time.Hour)))

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", "test")

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", "test")

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

							gitHistoryBasedCleanupCheck(ctx, imageName, 1, 0)

							setLastCommitCommitterWhen(ctx, time.Now().Add(-(23 * time.Hour)))

							utils.RunSucceedCommand(
								ctx,
								SuiteData.TestDirPath,
								"git",
								"push", "--force",
							)

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

							gitHistoryBasedCleanupCheck(ctx, imageName, 1, 1)
						})

						It("should remove image by references.limit.in", func(ctx SpecContext) {
							SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "4")

							setLastCommitCommitterWhen(ctx, time.Now().Add(-(13 * time.Hour)))

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", "test")

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", "test")

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

							gitHistoryBasedCleanupCheck(ctx, imageName, 1, 0)

							setLastCommitCommitterWhen(ctx, time.Now().Add(-(11 * time.Hour)))

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "--force")

							utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

							gitHistoryBasedCleanupCheck(ctx, imageName, 1, 1)
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

							BeforeEach(func(ctx SpecContext) {
								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", ref1)

								setLastCommitCommitterWhen(ctx, time.Now().Add(-(13 * time.Hour)))

								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", ref1)

								SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")
								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

								stageID1 = resultStageID(ctx, imageName)
								_ = stageID1

								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", ref2)

								setLastCommitCommitterWhen(ctx, time.Now().Add(-(11 * time.Hour)))

								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", ref2)

								SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "2")
								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

								stageID2 = resultStageID(ctx, imageName)

								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", ref3)

								setLastCommitCommitterWhen(ctx, time.Now())

								SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "3")
								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

								stageID3 = resultStageID(ctx, imageName)

								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", ref3)
							})

							It("should remove image by references.limit.in OR references.limit.last", func(ctx SpecContext) {
								SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "5")
								gitHistoryBasedCleanupCheck(ctx, imageName, 3, 2, func(imageMetadata map[string][]string) {
									Expect(imageMetadata).Should(HaveKey(stageID2))
									Expect(imageMetadata).Should(HaveKey(stageID3))
								})
							})

							It("should remove image by references.limit.in AND references.limit.last", func(ctx SpecContext) {
								SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "6")
								gitHistoryBasedCleanupCheck(ctx, imageName, 3, 1, func(imageMetadata map[string][]string) {
									Expect(imageMetadata).Should(HaveKey(stageID3))
								})
							})
						})

						Context("imagesPerReference.*", func() {
							var (
								stageID1 string
								stageID2 string
								stageID3 string
							)

							BeforeEach(func(ctx SpecContext) {
								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", "test")

								setLastCommitCommitterWhen(ctx, time.Now().Add(-(13 * time.Hour)))

								SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")
								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

								stageID1 = resultStageID(ctx, imageName)
								_ = stageID1

								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "--allow-empty", "-m", "+")

								setLastCommitCommitterWhen(ctx, time.Now().Add(-(11 * time.Hour)))

								SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "2")
								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

								stageID2 = resultStageID(ctx, imageName)

								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "--allow-empty", "-m", "+")

								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", "test")

								SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "3")
								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

								stageID3 = resultStageID(ctx, imageName)
							})

							It("should remove image by imagesPerReference.in OR imagesPerReference.last", func(ctx SpecContext) {
								SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "7")
								gitHistoryBasedCleanupCheck(ctx, imageName, 3, 2, func(imageMetadata map[string][]string) {
									Expect(imageMetadata).Should(HaveKey(stageID2))
									Expect(imageMetadata).Should(HaveKey(stageID3))
								})
							})

							It("should remove image by imagesPerReference.in AND imagesPerReference.last", func(ctx SpecContext) {
								SuiteData.Stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "8")
								gitHistoryBasedCleanupCheck(ctx, imageName, 3, 1, func(imageMetadata map[string][]string) {
									Expect(imageMetadata).Should(HaveKey(stageID3))
								})
							})
						})
					})

					Context("images metadata cleanup", func() {
						When("one content digest", func() {
							metaImagesCheckFunc := func(ctx context.Context, before, after int, afterExtraChecks ...func(commits []string)) {
								imageMetadata := ImageMetadata(ctx, imageName)
								for _, commitList := range imageMetadata {
									Expect(commitList).Should(HaveLen(before))
								}

								gitHistoryBasedCleanupCheck(ctx, imageName, 1, 1, func(imageMetadata map[string][]string) {
									for _, commitList := range imageMetadata {
										Expect(commitList).Should(HaveLen(after))

										for _, check := range afterExtraChecks {
											check(commitList)
										}
									}
								})
							}

							It("should keep image metadata only for the latest commit (one branch)", func(ctx SpecContext) {
								for i := 0; i < 3; i++ {
									utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "--allow-empty", "-m", "+")

									utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")
								}

								utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", branchName)

								metaImagesCheckFunc(ctx, 3, 1, func(commits []string) {
									Expect(commits).Should(ContainElement(utils.GetHeadCommit(ctx, SuiteData.TestDirPath)))
								})
							})

							It("should keep all image metadata (several branches)", func(ctx SpecContext) {
								for i := 0; i < 3; i++ {
									utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "--allow-empty", "-m", "+")

									utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

									branch := fmt.Sprintf("test_%d", i)
									utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", branch)

									utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", branch)
								}

								metaImagesCheckFunc(ctx, 3, 3)
							})
						})
					})
				})

				Context("custom tags", func() {
					It("should remove custom tag associated with the deleted stage", func(ctx SpecContext) {
						customTag1 := "tag1"
						customTag2 := "tag2"

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build", "--add-custom-tag", fmt.Sprintf(customTagValueFormat, customTag1))

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "push", "--set-upstream", "origin", branchName)

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "--allow-empty", "-m", "test")

						SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "REBUILD")

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build", "--add-custom-tag", fmt.Sprintf(customTagValueFormat, customTag2))

						customTags := CustomTags(ctx)
						Expect(customTags).Should(ContainElement(fmt.Sprintf(customTagValueFormat, customTag1)))
						Expect(customTags).Should(ContainElement(fmt.Sprintf(customTagValueFormat, customTag2)))
						Expect(len(CustomTagsMetadataList(ctx))).Should(Equal(2))

						utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "cleanup")

						customTags = CustomTags(ctx)
						Expect(customTags).Should(ContainElement(fmt.Sprintf(customTagValueFormat, customTag1)))
						Expect(customTags).ShouldNot(ContainElement(fmt.Sprintf(customTagValueFormat, customTag2)))
						Expect(len(CustomTagsMetadataList(ctx))).Should(Equal(1))
					})
				})
			})
		})
	}
})

func resultStageID(ctx context.Context, imageName string) string {
	res := utils.SucceedCommandOutputString(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "stage", "image", imageName)

	parts := strings.Split(strings.TrimSpace(res), ":")
	return parts[len(parts)-1]
}

func cleanupBeforeEachBase(ctx context.Context) {
	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "init", "--bare", "remote.git")

	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "init")

	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "remote", "add", "origin", "remote.git")

	utils.RunSucceedCommand(
		ctx,
		SuiteData.TestDirPath,
		"git",
		"add", "werf*.yaml",
	)

	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "-m", "Initial commit")

	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "checkout", "-b", branchName)
}

func gitHistoryBasedCleanupCheck(ctx context.Context, imageName string, expectedNumberOfMetadataTagsBefore, expectedNumberOfMetadataTagsAfter int, afterCleanupChecks ...func(map[string][]string)) {
	Expect(ImageMetadata(ctx, imageName)).Should(HaveLen(expectedNumberOfMetadataTagsBefore))

	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "cleanup")

	if SuiteData.TestImplementation != docker_registry.QuayImplementationName {
		imageMetadata := ImageMetadata(ctx, imageName)
		Expect(imageMetadata).Should(HaveLen(expectedNumberOfMetadataTagsAfter))

		for _, check := range afterCleanupChecks {
			check(imageMetadata)
		}
	}
}

func setLastCommitCommitterWhen(ctx context.Context, newDate time.Time) {
	_, _ = utils.RunCommandWithOptions(ctx, SuiteData.TestDirPath, "git", []string{"commit", "--amend", "--allow-empty", "--no-edit", "--date", newDate.Format(time.RFC3339)}, utils.RunCommandOptions{ShouldSucceed: true, ExtraEnv: []string{fmt.Sprintf("GIT_COMMITTER_DATE=%s", newDate.Format(time.RFC3339))}})
}
