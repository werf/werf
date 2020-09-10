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

var _ = forEachDockerRegistryImplementation("cleaning images", func() {
	for _, werfCommand := range [][]string{
		{"images", "cleanup"},
		{"cleanup"},
	} {
		description := strings.Join(werfCommand, " ") + " command"
		werfCommand := werfCommand

		Describe(description, func() {
			// Context("when deployed images in kubernetes are not taken into account", func()
			Context("git history based cleanup", func() {
				BeforeEach(func() {
					stubs.SetEnv("WERF_GIT_HISTORY_BASED_CLEANUP", "1")
					stubs.SetEnv("WERF_WITHOUT_KUBE", "1")
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

					It("should remove image by references.limit.in OR references.limit.last", func() {
						stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "5")
						imagesCleanupCheck(
							werfCommand,
							3,
							2,
							func(tags []string) {
								Ω(tags).Should(ContainElement(imagesRepo.ImageRepositoryTag("image", tag2)))
								Ω(tags).Should(ContainElement(imagesRepo.ImageRepositoryTag("image", tag3)))
							},
						)
					})

					It("should remove image by references.limit.in AND references.limit.last", func() {
						stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "6")
						imagesCleanupCheck(
							werfCommand,
							3,
							1,
							func(tags []string) {
								Ω(tags).Should(ContainElement(imagesRepo.ImageRepositoryTag("image", tag3)))
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

					It("should remove image by imagesPerReference.in OR imagesPerReference.last", func() {
						stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "7")
						imagesCleanupCheck(
							werfCommand,
							3,
							2,
							func(tags []string) {
								Ω(tags).Should(ContainElement(imagesRepo.ImageRepositoryTag("image", tag2)))
								Ω(tags).Should(ContainElement(imagesRepo.ImageRepositoryTag("image", tag3)))
							},
						)
					})

					It("should remove image by imagesPerReference.in AND imagesPerReference.last", func() {
						stubs.SetEnv("CLEANUP_POLICY_SET_NUMBER", "8")
						imagesCleanupCheck(
							werfCommand,
							3,
							1,
							func(tags []string) {
								Ω(tags).Should(ContainElement(imagesRepo.ImageRepositoryTag("image", tag3)))
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

	if testImplementation != docker_registry.QuayImplementationName {
		tags = imagesRepoAllImageRepoTags("image")
		Ω(tags).Should(HaveLen(expectedNumberOfTagsAfter))
	}

	if resultTagsChecks != nil {
		for _, check := range resultTagsChecks {
			check(tags)
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
