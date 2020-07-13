package cleanup_test

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/testing/utils"
)

var _ = forEachDockerRegistryImplementation("cleaning images", func() {
	for _, werfCommand := range [][]string{
		{"images", "cleanup"},
		{"cleanup"},
	} {
		description := strings.Join(werfCommand, " ") + " command"
		werfCommand := werfCommand

		Describe(description, func() {
			Context("when deployed images in kubernetes are not taken into account", func() {
				BeforeEach(func() {
					stubs.SetEnv("WERF_WITHOUT_KUBE", "1")
				})

				BeforeEach(func() {
					utils.CopyIn(utils.FixturePath("default"), testDirPath)
					imagesCleanupBeforeEachBase()
				})

				It("should work properly with non-existent/empty images repo", func() {
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						werfCommand...,
					)
				})

				Context("git branch strategy", func() {
					var testBranch = "branchA"

					It("should remove image that is associated with local branch", func() {
						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"checkout", "-b", testBranch,
						)

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build-and-publish", "--tag-git-branch", testBranch,
						)

						tags := imagesRepoAllImageRepoTags("image")
						Ω(tags).Should(ContainElement(imagesRepo.ImageRepositoryTag("image", testBranch)))

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfCommand...,
						)

						if testImplementation != docker_registry.QuayImplementationName {
							tags = imagesRepoAllImageRepoTags("image")
							Ω(tags).ShouldNot(ContainElement(imagesRepo.ImageRepositoryTag("image", testBranch)))
						}
					})

					It("should remove image that is associated with deleted remote branch", func() {
						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"checkout", "-b", testBranch,
						)

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"push", "--set-upstream", "origin", testBranch,
						)

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build-and-publish", "--tag-git-branch", testBranch,
						)

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfCommand...,
						)

						tags := imagesRepoAllImageRepoTags("image")
						Ω(tags).Should(ContainElement(imagesRepo.ImageRepositoryTag("image", testBranch)))

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"push", "origin", "--delete", testBranch,
						)

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfCommand...,
						)

						if testImplementation != docker_registry.QuayImplementationName {
							tags = imagesRepoAllImageRepoTags("image")
							Ω(tags).ShouldNot(ContainElement(imagesRepo.ImageRepositoryTag("image", testBranch)))
						}
					})
				})

				Context("git tag strategy", func() {
					var testTag = "tagA"

					It("should remove image that is associated with deleted tag", func() {
						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"tag", testTag,
						)

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build-and-publish", "--tag-git-tag", testTag,
						)

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfCommand...,
						)

						tags := imagesRepoAllImageRepoTags("image")
						Ω(tags).Should(ContainElement(imagesRepo.ImageRepositoryTag("image", testTag)))

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"tag", "-d", testTag,
						)

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfCommand...,
						)

						if testImplementation != docker_registry.QuayImplementationName {
							tags = imagesRepoAllImageRepoTags("image")
							Ω(tags).ShouldNot(ContainElement(imagesRepo.ImageRepositoryTag("image", testTag)))
						}
					})

					It("should remove image by expiry days policy (WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS)", func() {
						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"tag", testTag,
						)

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build-and-publish", "--tag-git-tag", testTag,
						)

						tags := imagesRepoAllImageRepoTags("image")
						Ω(tags).Should(HaveLen(1))

						werfArgs := append(werfCommand, "--git-tag-strategy-expiry-days", "0")
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfArgs...,
						)

						if testImplementation != docker_registry.QuayImplementationName {
							tags = imagesRepoAllImageRepoTags("image")
							Ω(tags).Should(HaveLen(0))
						}
					})

					It("should remove image by limit policy (WERF_GIT_TAG_STRATEGY_LIMIT)", func() {
						for _, tag := range []string{"tagA", "tagB", "tagC"} {
							utils.RunSucceedCommand(
								testDirPath,
								"git",
								"tag", tag,
							)

							utils.RunSucceedCommand(
								testDirPath,
								werfBinPath,
								"build-and-publish", "--tag-git-tag", tag,
							)
						}

						tags := imagesRepoAllImageRepoTags("image")
						Ω(tags).Should(HaveLen(3))

						werfArgs := append(werfCommand, "--git-tag-strategy-limit", "1")
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfArgs...,
						)

						if testImplementation != docker_registry.QuayImplementationName {
							tags = imagesRepoAllImageRepoTags("image")
							Ω(tags).Should(HaveLen(1))
						}
					})
				})

				Context("git commit strategy", func() {
					It("should remove image that is associated with non-existent commit", func() {
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build-and-publish", "--tag-git-commit", "8a99331ce0f918b411423223f4060e9688e03f6a",
						)

						tags := imagesRepoAllImageRepoTags("image")
						Ω(tags).Should(HaveLen(1))

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfCommand...,
						)

						if testImplementation != docker_registry.QuayImplementationName {
							tags = imagesRepoAllImageRepoTags("image")
							Ω(tags).Should(HaveLen(0))
						}
					})

					It("should not remove image that is associated with commit", func() {
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

						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfCommand...,
						)

						tags := imagesRepoAllImageRepoTags("image")
						Ω(tags).Should(ContainElement(imagesRepo.ImageRepositoryTag("image", commit)))
					})

					It("should remove image by expiry days policy (WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS)", func() {
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

						tags := imagesRepoAllImageRepoTags("image")
						Ω(tags).Should(HaveLen(1))

						werfArgs := append(werfCommand, "--git-commit-strategy-expiry-days", "0")
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfArgs...,
						)

						if testImplementation != docker_registry.QuayImplementationName {
							tags = imagesRepoAllImageRepoTags("image")
							Ω(tags).Should(HaveLen(0))
						}
					})

					It("should remove image by limit policy (WERF_GIT_COMMIT_STRATEGY_LIMIT)", func() {
						amount := 4
						for i := 0; i < amount; i++ {
							utils.RunSucceedCommand(
								testDirPath,
								"git",
								"commit", "--allow-empty", "--allow-empty-message", "-m", "",
							)

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
						}

						tags := imagesRepoAllImageRepoTags("image")
						Ω(tags).Should(HaveLen(amount))

						werfArgs := append(werfCommand, "--git-commit-strategy-limit", "2")
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							werfArgs...,
						)

						if testImplementation != docker_registry.QuayImplementationName {
							tags = imagesRepoAllImageRepoTags("image")
							Ω(tags).Should(HaveLen(2))
						}
					})
				})
			})

			Context("git history based cleanup", func() {
				BeforeEach(func() {
					stubs.SetEnv("WERF_GIT_HISTORY_BASED_CLEANUP", "1")
				})

				BeforeEach(func() {
					utils.CopyIn(utils.FixturePath("git_history_based_cleanup"), testDirPath)
					imagesCleanupBeforeEachBase()
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
						"build-and-publish", "--tag-by-stages-signature",
					)

					imagesCleanupCheck(werfCommand, 1, 1)

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"push", "origin", "--delete", "test",
					)

					imagesCleanupCheck(werfCommand, 1, 0)
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
						"build-and-publish", "--tag-by-stages-signature",
					)

					imagesCleanupCheck(werfCommand, 1, 0)
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
						"build-and-publish", "--tag-by-stages-signature",
					)

					imagesCleanupCheck(werfCommand, 1, 0)
				})

				It("should remove image by imagesPerReference.publishedIn", func() {
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
						"build-and-publish", "--tag-by-stages-signature",
					)

					imagesCleanupCheck(werfCommand, 1, 0)

					setLastCommitCommitterWhen(time.Now().Add(-(23 * time.Hour)))

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"push", "--force",
					)

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build-and-publish", "--tag-by-stages-signature",
					)

					imagesCleanupCheck(werfCommand, 1, 1)
				})

				It("should remove image by references.limit.createdIn", func() {
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
						"build-and-publish", "--tag-by-stages-signature",
					)

					imagesCleanupCheck(werfCommand, 1, 0)

					setLastCommitCommitterWhen(time.Now().Add(-(11 * time.Hour)))

					utils.RunSucceedCommand(
						testDirPath,
						"git",
						"push", "--set-upstream", "--force",
					)

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build-and-publish", "--tag-by-stages-signature",
					)

					imagesCleanupCheck(werfCommand, 1, 1)
				})

				Context("references.limit.*", func() {
					const (
						tag1 = "test1"
						tag2 = "test2"
						tag3 = "test3"
					)

					BeforeEach(func() {
						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"checkout", "-b", tag1,
						)

						setLastCommitCommitterWhen(time.Now().Add(-(13 * time.Hour)))

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"push", "--set-upstream", "origin", tag1,
						)

						stubs.SetEnv("FROM_CACHE_VERSION", "1")
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build-and-publish", "--tag-custom", tag1,
						)

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"checkout", "-b", tag2,
						)

						setLastCommitCommitterWhen(time.Now().Add(-(11 * time.Hour)))

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"push", "--set-upstream", "origin", tag2,
						)

						stubs.SetEnv("FROM_CACHE_VERSION", "2")
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build-and-publish", "--tag-custom", tag2,
						)

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"checkout", "-b", tag3,
						)

						setLastCommitCommitterWhen(time.Now())

						stubs.SetEnv("FROM_CACHE_VERSION", "3")
						utils.RunSucceedCommand(
							testDirPath,
							werfBinPath,
							"build-and-publish", "--tag-custom", tag3,
						)

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"push", "--set-upstream", "origin", tag3,
						)
					})

					It("should remove image by references.limit.createdIn OR references.limit.last", func() {
						stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "5")
						imagesCleanupCheck(
							werfCommand,
							3,
							2,
							func(tags []string) {
								Ω(tags).Should(ContainElement(tag2))
								Ω(tags).Should(ContainElement(tag3))
							},
						)
					})

					It("should remove image by references.limit.createdIn AND references.limit.last", func() {
						stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "6")
						imagesCleanupCheck(
							werfCommand,
							3,
							1,
							func(tags []string) {
								Ω(tags).Should(ContainElement(tag3))
							},
						)
					})
				})

				Context("imagesPerReference.*", func() {
					const (
						tag1 = "test1"
						tag2 = "test2"
						tag3 = "test3"
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
							"build-and-publish", "--tag-custom", tag1,
						)

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
							"build-and-publish", "--tag-custom", tag2,
						)

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
							"build-and-publish", "--tag-custom", tag3,
						)
					})

					It("should remove image by imagesPerReference.publishedIn OR imagesPerReference.last", func() {
						stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "7")
						imagesCleanupCheck(
							werfCommand,
							3,
							2,
							func(tags []string) {
								Ω(tags).Should(ContainElement(tag2))
								Ω(tags).Should(ContainElement(tag3))
							},
						)
					})

					It("should remove image by imagesPerReference.publishedIn AND imagesPerReference.last", func() {
						stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "8")
						imagesCleanupCheck(
							werfCommand,
							3,
							1,
							func(tags []string) {
								Ω(tags).Should(ContainElement(tag3))
							},
						)
					})
				})
			})
		})
	}
})

func imagesCleanupBeforeEachBase() {
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

func imagesCleanupCheck(cleanupArgs []string, expectedNumberOfTagsBefore, expectedNumberOfTagsAfter int, resultTagsChecks ...func([]string)) {
	tags := imagesRepoAllImageRepoTags("image")
	Ω(tags).Should(HaveLen(expectedNumberOfTagsBefore))

	utils.RunSucceedCommand(
		testDirPath,
		werfBinPath,
		cleanupArgs...,
	)

	tags = imagesRepoAllImageRepoTags("image")
	Ω(tags).Should(HaveLen(expectedNumberOfTagsAfter))

	if resultTagsChecks != nil {
		for _, check := range resultTagsChecks {
			check(tags)
		}
	}
}

func setLastCommitCommitterWhen(newDate time.Time) {
	stubs.SetEnv("GIT_COMMITTER_DATE", newDate.Format(time.RFC3339))
	utils.RunSucceedCommand(
		testDirPath,
		"git",
		"commit", "--amend", "--allow-empty", "--no-edit", "--date", newDate.Format(time.RFC3339),
	)
}
