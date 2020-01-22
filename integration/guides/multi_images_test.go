package guides_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"

	"github.com/flant/werf/pkg/testing/utils"
	utilsDocker "github.com/flant/werf/pkg/testing/utils/docker"
	"github.com/flant/werf/pkg/testing/utils/net"
)

var _ = Describe("Advanced build/Multi images", func() {
	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	It("application should be built and checked", func() {
		utils.RunSucceedCommand(
			".",
			"git",
			"clone", "https://github.com/dockersamples/atsea-sample-shop-app.git", testDirPath,
		)

		utils.CopyIn(utils.FixturePath("multi_images"), testDirPath)

		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"build", "-s", ":local",
		)

		paymentGWContainerName := fmt.Sprintf("payment_gw_%s", utils.GetRandomString(10))
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d --name %s", paymentGWContainerName), "payment_gw",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(paymentGWContainerName) }()

		databaseContainerHostPort := net.GetFreeTCPHostPort()
		databaseContainerName := fmt.Sprintf("database_%s", utils.GetRandomString(10))
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p %d:5432 --name %s", databaseContainerHostPort, databaseContainerName), "database",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(databaseContainerName) }()

		appContainerHostPort := net.GetFreeTCPHostPort()
		appContainerName := fmt.Sprintf("app_%s", utils.GetRandomString(10))
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p %d:8080 --link %s:database --name %s", appContainerHostPort, databaseContainerName, appContainerName), "app",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(appContainerName) }()

		reverseProxyContainerHostPort80 := net.GetFreeTCPHostPort()
		reverseProxyContainerHostPort443 := net.GetFreeTCPHostPort()
		reverseProxyContainerName := fmt.Sprintf("reverse_proxy_%s", utils.GetRandomString(10))
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p %d:80 -p %d:443 --link %s:appserver --name %s", reverseProxyContainerHostPort80, reverseProxyContainerHostPort443, appContainerName, reverseProxyContainerName), "reverse_proxy",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(reverseProxyContainerName) }()

		url := fmt.Sprintf("http://localhost:%d/index.html", appContainerHostPort)
		waitTillHostReadyAndCheckResponseBody(
			url,
			360,
			"Atsea Shop",
		)
	})
})
