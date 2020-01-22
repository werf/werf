package guides_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"

	"github.com/flant/werf/pkg/testing/utils"
	utilsDocker "github.com/flant/werf/pkg/testing/utils/docker"
	"github.com/flant/werf/pkg/testing/utils/net"
)

var _ = Describe("Advanced build/Artifacts", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("artifacts"), testDirPath)
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

		containerHostPort := net.GetFreeTCPHostPort()
		containerName := fmt.Sprintf("go_booking_artifacts_%s", utils.GetRandomString(10))
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p %d:9000 --name %s", containerHostPort, containerName), "go-booking", "--", "/app/run.sh",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(containerName) }()

		url := fmt.Sprintf("http://localhost:%d", containerHostPort)
		waitTillHostReadyAndCheckResponseBody(
			url,
			net.DefaultWaitTillHostReadyToRespondMaxAttempts,
			"revel framework booking demo",
		)
	})
})
