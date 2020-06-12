package cleanup_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/docker_registry"

	"github.com/werf/werf/pkg/testing/utils"
)

var _ = forEachDockerRegistryImplementation("cleaning images", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("default"), testDirPath)

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
	})

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
		})
	}
})
