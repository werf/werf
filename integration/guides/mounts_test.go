package guides_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"

	"github.com/werf/werf/pkg/testing/utils"
	utilsDocker "github.com/werf/werf/pkg/testing/utils/docker"
)

var _ = Describe("Advanced build/Mounts", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("mounts"), testDirPath)
	})

	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	It("application should be built and checked", func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"build", "-s", ":local",
		)

		containerName := fmt.Sprintf("go_booking_mounts_%s", utils.GetRandomString(10))
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p :9000 --name %s", containerName), "go-booking", "--", "/app/run.sh",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(containerName) }()

		url := fmt.Sprintf("http://localhost:%s", utilsDocker.ContainerHostPort(containerName, "9000/tcp"))
		waitTillHostReadyAndCheckResponseBody(
			url,
			utils.DefaultWaitTillHostReadyToRespondMaxAttempts,
			"revel framework booking demo",
		)
	})
})
