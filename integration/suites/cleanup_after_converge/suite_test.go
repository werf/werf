package cleanup_with_k8s_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Cleanup after converge suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"git", "docker"},
	RequiredSuiteEnvs: []string{
		"WERF_TEST_K8S_DOCKER_REGISTRY",
	},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData = struct {
	suite_init.SuiteData
	StagesStorage     storage.PrimaryStagesStorage
	ContainerRegistry docker_registry.Interface
}{}

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupK8sDockerRegistry(suite_init.NewK8sDockerRegistryData(SuiteData.ProjectNameData, SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())
)

var _ = BeforeEach(func(ctx SpecContext) {
	SuiteData.StagesStorage = utils.NewStagesStorage(ctx, SuiteData.K8sDockerRegistryRepo, "default", docker_registry.DockerRegistryOptions{})

	containerRegistry, err := docker_registry.NewDockerRegistry(ctx, SuiteData.K8sDockerRegistryRepo, "", docker_registry.DockerRegistryOptions{})
	Expect(err).ShouldNot(HaveOccurred())

	SuiteData.ContainerRegistry = containerRegistry
})

func StagesCount() int {
	return utils.StagesCount(context.Background(), SuiteData.StagesStorage)
}

func ImportMetadataIDs() []string {
	return utils.ImportMetadataIDs(context.Background(), SuiteData.StagesStorage)
}
