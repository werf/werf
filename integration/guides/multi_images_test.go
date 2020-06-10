package guides_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"

	"github.com/werf/werf/pkg/testing/utils"
	utilsDocker "github.com/werf/werf/pkg/testing/utils/docker"
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

		databaseContainerName := fmt.Sprintf("database_%s", utils.GetRandomString(10))
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p :5432 --name %s", databaseContainerName), "database",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(databaseContainerName) }()

		appContainerName := fmt.Sprintf("app_%s", utils.GetRandomString(10))
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p :8080 --link %s:database --name %s", databaseContainerName, appContainerName), "app",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(appContainerName) }()

		reverseProxyContainerName := fmt.Sprintf("reverse_proxy_%s", utils.GetRandomString(10))
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local", "--docker-options", fmt.Sprintf("-d -p :80 -p :443 --link %s:appserver --name %s", appContainerName, reverseProxyContainerName), "reverse_proxy",
		)
		defer func() { utilsDocker.ContainerStopAndRemove(reverseProxyContainerName) }()

		url := fmt.Sprintf("http://localhost:%s/index.html", utilsDocker.ContainerHostPort(appContainerName, "8080/tcp"))
		waitTillHostReadyAndCheckResponseBody(
			url,
			360,
			"Atsea Shop",
		)
	})
})
