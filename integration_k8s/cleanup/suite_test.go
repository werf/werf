package cleanup_test

import (
	"context"
	"testing"

	"github.com/werf/werf/integration/suite_init"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/werf/werf/integration/utils"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"
)

const imageName = ""

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Cleanup suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"git", "docker"},
	RequiredSuiteEnvs: []string{
		"WERF_TEST_K8S_DOCKER_REGISTRY",
		"WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME",
		"WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD",
	},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData = struct {
	suite_init.SuiteData
	StagesStorage storage.StagesStorage
}{}

var _ = SuiteData.StubsData.Setup()
var _ = SuiteData.SynchronizedSuiteCallbacksData.Setup()
var _ = SuiteData.WerfBinaryData.Setup(&SuiteData.SynchronizedSuiteCallbacksData)
var _ = SuiteData.ProjectNameData.Setup(&SuiteData.StubsData)
var _ = SuiteData.K8sDockerRegistryData.Setup(&SuiteData.ProjectNameData, &SuiteData.StubsData)
var _ = SuiteData.TmpDirData.Setup()

var _ = BeforeEach(func() {
	SuiteData.StagesStorage = utils.NewStagesStorage(SuiteData.K8sDockerRegistryRepo, "default", docker_registry.DockerRegistryOptions{})
})

func StagesCount() int {
	return utils.StagesCount(context.Background(), SuiteData.StagesStorage)
}

func ImageMetadata(imageName string) map[string][]string {
	return utils.ImageMetadata(context.Background(), SuiteData.StagesStorage, imageName)
}

var _ = ginkgo.AfterEach(func() {
	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		SuiteData.WerfBinPath,
		"purge", "--force",
	)
})
