package cleanup_with_k8s_test

import (
	"context"
	"testing"

	"github.com/werf/werf/integration/pkg/suite_init"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"
)

const imageName = ""

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Cleanup after converge suite", suite_init.TestSuiteEntrypointFuncOptions{
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

var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
var _ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
var _ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
var _ = SuiteData.SetupK8sDockerRegistry(suite_init.NewK8sDockerRegistryData(SuiteData.ProjectNameData, SuiteData.StubsData))
var _ = SuiteData.SetupTmp(suite_init.NewTmpDirData())

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
