// +build integration

package guides_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"

	"github.com/flant/werf/integration/utils"
	utilsDocker "github.com/flant/werf/integration/utils/docker"
)

var _ = Describe("Guide/Advanced build/First application", func() {
	var testDirPath string
	var testName = "first_application"

	AfterEach(func() {
		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	for _, elm := range []string{"shell", "ansible"} {
		boundedBuilder := elm

		It(fmt.Sprintf("%s application should be built, checked and published", boundedBuilder), func() {
			testDirPath = tmpPath(testName, boundedBuilder)

			utils.RunCommand(
				".",
				"git",
				"clone", "https://github.com/symfony/symfony-demo.git", testDirPath,
			)

			utils.CopyIn(fixturePath(testName, boundedBuilder), testDirPath)

			utils.RunCommand(
				testDirPath,
				werfBinPath,
				"build", "-s", ":local",
			)

			containerHostPort := utils.GetFreeTCPHostPort()
			containerName := fmt.Sprintf("symfony_demo_%s_%s", boundedBuilder, utils.GetRandomString(10))
			utils.RunCommand(
				testDirPath,
				werfBinPath,
				"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p %d:8000 --name %s", containerHostPort, containerName), "--", "/app/start.sh",
			)
			defer func() { utilsDocker.ContainerStopAndRemove(containerName) }()

			url := fmt.Sprintf("http://localhost:%d", containerHostPort)
			waitTillHostReadyAndCheckResponseBody(
				url,
				utils.DefaultWaitTillHostReadyToRespondMaxAttempts,
				"Symfony Demo application",
			)

			registry, registryContainerName := utilsDocker.LocalDockerRegistryRun()
			defer func() { utilsDocker.ContainerStopAndRemove(registryContainerName) }()

			registryRepositoryName := containerName
			utils.RunCommand(
				testDirPath,
				werfBinPath,
				"publish", "-s", ":local", "-i", fmt.Sprintf("%s/%s", registry, registryRepositoryName), "--tag-custom", "test",
			)
		})
	}
})
