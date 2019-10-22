// +build integration

package cleanup_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/integration/utils"
	utilsDocker "github.com/flant/werf/integration/utils/docker"
)

var _ = Describe("images purge command", func() {
	var testDirPath string
	var registry, registryRepository, registryContainerName string
	var testName = "images_purge"

	BeforeEach(func() {
		testDirPath = tmpPath(testName)
		utils.CreateSimpleWerfYaml(testDirPath)

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

		registry, registryContainerName = utilsDocker.LocalDockerRegistryRun()
		registryRepository = strings.Join([]string{registry, "test"}, "/")

		Ω(os.Setenv("WERF_IMAGES_REPO", registryRepository)).Should(Succeed())
		Ω(os.Setenv("WERF_STAGES_STORAGE", ":local")).Should(Succeed())
	})

	AfterEach(func() {
		utilsDocker.ContainerStopAndRemove(registryContainerName)

		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	Context("when deployed images in kubernetes are not taken in account", func() {
		BeforeEach(func() {
			Ω(os.Setenv("WERF_WITHOUT_KUBE", "1")).Should(Succeed())
		})

		It("should work properly with non-existent registry repository (registry exists)", func() {
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"images", "purge",
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

				out := utils.SucceedCommandOutput(
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

			tags := utils.RegistryRepositoryList(registryRepository)
			Ω(tags).Should(HaveLen(amount))

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"images", "purge",
			)

			tags = utils.RegistryRepositoryList(registryRepository)
			Ω(tags).Should(HaveLen(0))
		})

		It("should not remove images built without werf", func() {
			Ω(utilsDocker.CliPull("alpine")).Should(Succeed(), "docker pull")
			Ω(utilsDocker.CliTag("alpine", registryRepository)).Should(Succeed(), "docker tag")
			defer func() { Ω(utilsDocker.CliRmi(registryRepository)).Should(Succeed(), "docker rmi") }()

			Ω(utilsDocker.CliPush(registryRepository)).Should(Succeed(), "docker push")

			tags := utils.RegistryRepositoryList(registryRepository)
			Ω(tags).Should(HaveLen(1))

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"images", "purge",
			)

			tags = utils.RegistryRepositoryList(registryRepository)
			Ω(tags).Should(HaveLen(1))
		})
	})
})
