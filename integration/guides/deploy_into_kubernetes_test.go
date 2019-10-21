// +build integration_k8s

package guides_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/integration/utils"
)

var _ = Describe("Guide/Getting started", func() {
	var testDirPath string
	var testName = "deploy_into_kubernetes"

	requiredSuiteEnvs = append(
		requiredSuiteEnvs,
		"WERF_TEST_K8S_DOCKER_REGISTRY",
		"WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME",
		"WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD",
	)

	AfterEach(func() {
		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	It("application should be built, published and deployed", func() {
		testDirPath = tmpPath(testName)

		projectName := fmt.Sprintf("deploy-into-kubernetes-%s", utils.GetRandomString(10))
		Ω(os.Setenv("WERF_TEST_K8S_PROJECT_NAME", projectName)).Should(Succeed(), "set env")

		utils.CopyIn(fixturePath(testName), testDirPath)

		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"build", "-s", ":local",
		)

		imagesRepo := fmt.Sprintf("%s/%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), projectName)
		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"publish", "-s", ":local", "-i", imagesRepo, "--tag-custom", "test",
		)

		werfDeployArgs := []string{
			"deploy",
			"-s", ":local",
			"-i", imagesRepo,
			"--tag-custom", "test",
			"--env", "test",
			"--set", fmt.Sprintf("imageCredentials.registry=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY")),
			"--set", fmt.Sprintf("imageCredentials.username=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME")),
			"--set", fmt.Sprintf("imageCredentials.password=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD")),
		}
		utils.RunCommand(
			testDirPath,
			werfBinPath,
			werfDeployArgs...,
		)

		utils.RunCommand(
			testDirPath,
			werfBinPath,
			"dismiss", "--env", "test",
		)
	})
})
