package suite_init

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	"github.com/prashantv/gostub"
)

type K8sDockerRegistryData struct {
	K8sDockerRegistryRepo string
}

func NewK8sDockerRegistryData(projectNameData *ProjectNameData, stubsData *StubsData) *K8sDockerRegistryData {
	data := &K8sDockerRegistryData{}
	SetupK8sDockerRegistryRepo(&data.K8sDockerRegistryRepo, &projectNameData.ProjectName, stubsData.Stubs)
	return data
}

func SetupK8sDockerRegistryRepo(repo, projectName *string, stubs *gostub.Stubs) bool {
	return BeforeEach(func() {
		*repo = fmt.Sprintf("%s/%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), *projectName)
		stubs.SetEnv("WERF_REPO", *repo)
	})
}
