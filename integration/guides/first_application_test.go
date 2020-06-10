package guides_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"

	"github.com/werf/werf/pkg/testing/utils"
	utilsDocker "github.com/werf/werf/pkg/testing/utils/docker"
)

var _ = Describe("Advanced build/First application", func() {
	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	for _, builder := range []string{"shell", "ansible"} {
		boundedBuilder := builder

		It(fmt.Sprintf("%s application should be built, checked and published", boundedBuilder), func() {
			utils.RunSucceedCommand(
				".",
				"git",
				"clone", "https://github.com/symfony/symfony-demo.git", testDirPath,
			)

			utils.CopyIn(utils.FixturePath("first_application", boundedBuilder), testDirPath)

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"build", "-s", ":local",
			)

			containerName := fmt.Sprintf("symfony_demo_%s_%s", boundedBuilder, utils.GetRandomString(10))
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p :8000 --name %s", containerName), "--", "/app/start.sh",
			)
			defer func() { utilsDocker.ContainerStopAndRemove(containerName) }()

			url := fmt.Sprintf("http://localhost:%s", utilsDocker.ContainerHostPort(containerName, "8000/tcp"))
			waitTillHostReadyAndCheckResponseBody(
				url,
				utils.DefaultWaitTillHostReadyToRespondMaxAttempts,
				"Symfony Demo application",
			)

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"publish", "-s", ":local", "-i", registryProjectRepository, "--tag-custom", "test",
			)
		})
	}
})
