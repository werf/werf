// +build integration

package guides_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"

	"github.com/flant/werf/integration/utils"
	utilsDocker "github.com/flant/werf/integration/utils/docker"
)

var _ = Describe("Guide/Advanced build/Multi images", func() {
	var testDirPath string
	var testName = "multi_images"

	AfterEach(func() {
		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	It("application should be built and checked", func() {
		testDirPath = tmpPath(testName)

		utils.RunCommand(
			".",
			"git",
			"clone", "https://github.com/dockersamples/atsea-sample-shop-app.git", testDirPath,
		)

		utils.CopyIn(fixturePath(testName), testDirPath)

		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"build", "-s", ":local",
		)

		paymentGWContainerName := fmt.Sprintf("payment_gw_%s", utils.GetRandomString(10))
		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d --name %s", paymentGWContainerName), "payment_gw",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(paymentGWContainerName) }()

		databaseContainerHostPort := utils.GetFreeTCPHostPort()
		databaseContainerName := fmt.Sprintf("database_%s", utils.GetRandomString(10))
		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p %d:5432 --name %s", databaseContainerHostPort, databaseContainerName), "database",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(databaseContainerName) }()

		appContainerHostPort := utils.GetFreeTCPHostPort()
		appContainerName := fmt.Sprintf("app_%s", utils.GetRandomString(10))
		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p %d:8080 --link %s:database --name %s", appContainerHostPort, databaseContainerName, appContainerName), "app",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(appContainerName) }()

		reverseProxyContainerHostPort80 := utils.GetFreeTCPHostPort()
		reverseProxyContainerHostPort443 := utils.GetFreeTCPHostPort()
		reverseProxyContainerName := fmt.Sprintf("reverse_proxy_%s", utils.GetRandomString(10))
		utils.RunCommand(
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
