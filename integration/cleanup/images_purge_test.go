package cleanup_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/docker_registry"

	"github.com/werf/werf/pkg/testing/utils"
	utilsDocker "github.com/werf/werf/pkg/testing/utils/docker"
)

var _ = forEachDockerRegistryImplementation("purging images", func() {
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
		{"images", "purge"},
		{"purge"},
	} {
		description := strings.Join(werfCommand, " ") + " command"
		werfCommand := werfCommand

		Describe(description, func() {
			It("should work properly with non-existent/empty images repo", func() {
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					werfCommand...,
				)
			})

			It("should remove images that are built with werf", func() {
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
						"build-and-publish", "--tag-custom", commit,
					)
				}

				if testImplementation != docker_registry.QuayImplementationName {
					tags := imagesRepoAllImageRepoTags("image")
					Ω(tags).Should(HaveLen(amount))
				}

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					werfCommand...,
				)

				if testImplementation != docker_registry.QuayImplementationName {
					tags := imagesRepoAllImageRepoTags("image")
					Ω(tags).Should(HaveLen(0))
				}
			})

			It("should not remove images that are built without werf", func() {
				Ω(utilsDocker.Pull("flant/werf-test:hello-world")).Should(Succeed(), "docker pull")
				Ω(utilsDocker.CliTag("flant/werf-test:hello-world", imagesRepo.ImageRepositoryName("image"))).Should(Succeed(), "docker tag")
				defer func() {
					Ω(utilsDocker.CliRmi(imagesRepo.ImageRepositoryName("image"))).Should(Succeed(), "docker rmi")
				}()

				Ω(utilsDocker.CliPush(imagesRepo.ImageRepositoryName("image"))).Should(Succeed(), "docker push")

				tags := imagesRepoAllImageRepoTags("image")
				Ω(tags).Should(HaveLen(1))

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					werfCommand...,
				)

				tags = imagesRepoAllImageRepoTags("image")
				Ω(tags).Should(HaveLen(1))
			})
		})
	}
})
