// +build integration

package guides_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"

	"github.com/flant/werf/integration/utils"
	utilsDocker "github.com/flant/werf/integration/utils/docker"
)

var _ = Describe("Guide/Getting started", func() {
	var testDirPath string
	var testName = "getting_started"

	AfterEach(func() {
		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	It("application should be built, checked and published", func() {
		testDirPath = tmpPath(testName)

		utils.RunCommand(
			".",
			"git",
			"clone", "https://github.com/dockersamples/linux_tweet_app.git", testDirPath,
		)

		utils.CopyIn(fixturePath(testName), testDirPath)

		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"build", "-s", ":local",
		)

		containerHostPort := utils.GetFreeTCPHostPort()
		containerName := fmt.Sprintf("getting_started_%s", utils.GetRandomString(10))

		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p %d:80 --name %s", containerHostPort, containerName),
		)
		defer func() { utilsDocker.ContainerStopAndRemove(containerName) }()

		url := fmt.Sprintf("http://localhost:%d", containerHostPort)
		waitTillHostReadyAndCheckResponseBody(
			url,
			utils.DefaultWaitTillHostReadyToRespondMaxAttempts,
			"Linux Tweet App!",
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
})
