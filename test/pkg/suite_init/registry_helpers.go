package suite_init

import (
	"fmt"

	"github.com/werf/werf/v2/test/pkg/utils"
)

const (
	TestK8sDockerRegistryEnv = "WERF_TEST_K8S_DOCKER_REGISTRY"
)

// TestRegistry returns registry address in form localhost:port
func TestRegistry() string {
	return utils.GetRequiredEnv(TestK8sDockerRegistryEnv)
}

// TestRepo returns full werf repo: localhost:port/project
func TestRepo(projectName string) string {
	return fmt.Sprintf("%s/%s", TestRegistry(), projectName)
}
