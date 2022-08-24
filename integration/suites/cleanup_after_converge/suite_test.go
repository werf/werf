package cleanup_with_k8s_test

import (
	"context"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/test/pkg/suite_init"
	"github.com/werf/werf/test/pkg/utils"
)

const imageName = "backend"

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

var _ = BeforeEach(func() {
	SuiteData.StagesStorage = utils.NewStagesStorage(SuiteData.K8sDockerRegistryRepo, "default", docker_registry.DockerRegistryOptions{})

	containerRegistry, err := docker_registry.NewDockerRegistry(SuiteData.K8sDockerRegistryRepo, "", docker_registry.DockerRegistryOptions{})
	Ω(err).ShouldNot(HaveOccurred())

	SuiteData.ContainerRegistry = containerRegistry
})

func StagesCount() int {
	return utils.StagesCount(context.Background(), SuiteData.StagesStorage)
}

func ImageMetadata(imageName string) map[string][]string {
	return utils.ImageMetadata(context.Background(), SuiteData.StagesStorage, imageName)
}

func CustomTags() []string {
	tags, err := SuiteData.ContainerRegistry.Tags(context.Background(), SuiteData.K8sDockerRegistryRepo)
	Ω(err).ShouldNot(HaveOccurred())

	var result []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, customTagValuePrefix) {
			result = append(result, tag)
		}
	}

	return result
}

func CustomTagsMetadataList() []*storage.CustomTagMetadata {
	return utils.CustomTagsMetadataList(context.Background(), SuiteData.StagesStorage)
}
