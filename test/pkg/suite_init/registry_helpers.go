package suite_init

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
)

const (
	TestK8sDockerRegistryEnv = "WERF_TEST_K8S_DOCKER_REGISTRY"
)

// Labels naming the external resources a spec cannot run without, so that a run
// can be narrowed to what the host actually provides.
const (
	LabelNeedsRegistry = "needs-registry"
	LabelNeedsKube     = "needs-kube"
	LabelNeedsBuildah  = "needs-buildah"
)

// TestRegistry returns registry address in form localhost:port, skipping the
// current spec when no test registry is configured.
func TestRegistry() string {
	registry := os.Getenv(TestK8sDockerRegistryEnv)
	if registry == "" {
		Skip(fmt.Sprintf("%s is not set", TestK8sDockerRegistryEnv))
	}

	return registry
}

// TestRepo returns full werf repo: localhost:port/project
func TestRepo(projectName string) string {
	return fmt.Sprintf("%s/%s", TestRegistry(), projectName)
}
