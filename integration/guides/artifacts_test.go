package guides_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"

	"github.com/werf/werf/pkg/testing/utils"
	utilsDocker "github.com/werf/werf/pkg/testing/utils/docker"
)

var _ = Describe("Advanced build/Artifacts", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("artifacts"), testDirPath)
	})

	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "--force",
		)
	})

	It("application should be built and checked", func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"build",
		)

		containerName := fmt.Sprintf("gowebapp_build_artifacts_%s", utils.GetRandomString(10))
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"run", "--docker-options", fmt.Sprintf("-d -p :80 --name %s", containerName), "gowebapp", "--", "/app/gowebapp",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(containerName) }()

		url := fmt.Sprintf("http://localhost:%s", utilsDocker.ContainerHostPort(containerName, "80/tcp"))
		waitTillHostReadyAndCheckResponseBody(
			url,
			utils.DefaultWaitTillHostReadyToRespondMaxAttempts,
			"Go Web App",
		)
	})
})
