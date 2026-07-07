package cleanup_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
)

const (
	imageName = "image"

	customTagValuePrefix = "user-custom-tag-"
	customTagValueFormat = "user-custom-tag-%v"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Cleanup suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"git", "docker"},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData struct {
	suite_init.SuiteData
	TestImplementation string
	RegistryStorage      storage.RegistryStorage
	ContainerRegistry  docker_registry.Interface
}

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())
	_ = SuiteData.SetupContainerRegistryPerImplementation(suite_init.NewContainerRegistryPerImplementationData(SuiteData.SynchronizedSuiteCallbacksData, true))
)

func perImplementationBeforeEach(implementationName string) func(ctx SpecContext) {
	return func(ctx SpecContext) {
		Expect(werf.Init(SuiteData.TmpDir, "")).To(Succeed())

		werfImplementationName := SuiteData.ContainerRegistryPerImplementation[implementationName].WerfImplementationName

		repo := fmt.Sprintf("%s/%s", SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryAddress, SuiteData.ProjectName)
		InitRegistryStorage(ctx, repo, werfImplementationName, SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryOptions)
		SuiteData.SetupRepo(ctx, repo, implementationName, SuiteData.StubsData)
		SuiteData.TestImplementation = implementationName

		containerRegistry, err := docker_registry.NewDockerRegistry(ctx, repo, werfImplementationName, docker_registry.DockerRegistryOptions{})
		Expect(err).ShouldNot(HaveOccurred())

		SuiteData.ContainerRegistry = containerRegistry
	}
}

func InitRegistryStorage(ctx context.Context, stagesStorageAddress, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) {
	SuiteData.RegistryStorage = utils.NewRegistryStorage(ctx, stagesStorageAddress, implementationName, dockerRegistryOptions)
}

func StagesCount(ctx context.Context) int {
	return utils.StagesCount(ctx, SuiteData.RegistryStorage)
}

func ManagedImagesCount(ctx context.Context) int {
	return utils.ManagedImagesCount(ctx, SuiteData.RegistryStorage)
}

func ImageMetadata(ctx context.Context, imageName string) map[string][]string {
	return utils.ImageMetadata(ctx, SuiteData.RegistryStorage, imageName)
}

func CustomTags(ctx context.Context) []string {
	tags, err := SuiteData.ContainerRegistry.Tags(ctx, SuiteData.RegistryStorage.String())
	Expect(err).ShouldNot(HaveOccurred())

	var result []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, customTagValuePrefix) {
			result = append(result, tag)
		}
	}

	return result
}

func CustomTagsMetadataList(ctx context.Context) []*storage.CustomTagMetadata {
	return utils.CustomTagsMetadataList(ctx, SuiteData.RegistryStorage)
}
