package guides_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"

	"github.com/flant/werf/pkg/testing/utils"
)

var _ = Describe("Deploy into kubernetes", func() {
	requiredSuiteEnvs = append(
		requiredSuiteEnvs,
		"WERF_TEST_K8S_DOCKER_REGISTRY",
		"WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME",
		"WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD",
	)

	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("deploy_into_kubernetes"), testDirPath)
	})

	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)

		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"dismiss", "--env", "test", "--with-namespace",
		)
	})

	It("application should be built, published and deployed", func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"build", "-s", ":local",
		)

		imagesRepo := fmt.Sprintf("%s/%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), utils.ProjectName())
		utils.RunSucceedCommand(
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
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			werfDeployArgs...,
		)
	})
})
