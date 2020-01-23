package cleanup_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
	utilsDocker "github.com/flant/werf/pkg/testing/utils/docker"
)

var _ = Describe("purging images", func() {
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

		stubs.SetEnv("WERF_IMAGES_REPO", registryProjectRepository)
		stubs.SetEnv("WERF_STAGES_STORAGE", ":local")
	})

	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	for _, werfArgs := range [][]string{
		{"images", "purge"},
		{"purge"},
	} {
		commandToCheck := strings.Join(werfArgs, " ") + " command"
		commandWerfArgs := werfArgs

		Describe(commandToCheck, func() {
			It("should work properly with non-existent registry repository (registry exists)", func() {
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					commandWerfArgs...,
				)
			})

			It("should remove images built with werf", func() {
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

				tags := utils.RegistryRepositoryList(registryProjectRepository)
				Ω(tags).Should(HaveLen(amount))

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					commandWerfArgs...,
				)

				tags = utils.RegistryRepositoryList(registryProjectRepository)
				Ω(tags).Should(HaveLen(0))
			})

			It("should not remove images built without werf", func() {
				Ω(utilsDocker.Pull("alpine")).Should(Succeed(), "docker pull")
				Ω(utilsDocker.CliTag("alpine", registryProjectRepository)).Should(Succeed(), "docker tag")
				defer func() { Ω(utilsDocker.CliRmi(registryProjectRepository)).Should(Succeed(), "docker rmi") }()

				Ω(utilsDocker.CliPush(registryProjectRepository)).Should(Succeed(), "docker push")

				tags := utils.RegistryRepositoryList(registryProjectRepository)
				Ω(tags).Should(HaveLen(1))

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					commandWerfArgs...,
				)

				tags = utils.RegistryRepositoryList(registryProjectRepository)
				Ω(tags).Should(HaveLen(1))
			})
		})
	}
})
