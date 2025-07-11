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
	StagesStorage      storage.PrimaryStagesStorage
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
		werfImplementationName := SuiteData.ContainerRegistryPerImplementation[implementationName].WerfImplementationName

		repo := fmt.Sprintf("%s/%s", SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryAddress, SuiteData.ProjectName)
		InitStagesStorage(repo, werfImplementationName, SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryOptions)
		SuiteData.SetupRepo(ctx, repo, implementationName, SuiteData.StubsData)
		SuiteData.TestImplementation = implementationName

		containerRegistry, err := docker_registry.NewDockerRegistry(repo, werfImplementationName, docker_registry.DockerRegistryOptions{})
		Expect(err).ShouldNot(HaveOccurred())

		SuiteData.ContainerRegistry = containerRegistry
	}
}

func InitStagesStorage(stagesStorageAddress, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) {
	SuiteData.StagesStorage = utils.NewStagesStorage(stagesStorageAddress, implementationName, dockerRegistryOptions)
}

func StagesCount(ctx context.Context) int {
	return utils.StagesCount(ctx, SuiteData.StagesStorage)
}

func ManagedImagesCount(ctx context.Context) int {
	return utils.ManagedImagesCount(ctx, SuiteData.StagesStorage)
}

func ImageMetadata(ctx context.Context, imageName string) map[string][]string {
	return utils.ImageMetadata(ctx, SuiteData.StagesStorage, imageName)
}

func ImportMetadataIDs(ctx context.Context) []string {
	return utils.ImportMetadataIDs(ctx, SuiteData.StagesStorage)
}

func CustomTags(ctx context.Context) []string {
	tags, err := SuiteData.ContainerRegistry.Tags(ctx, SuiteData.StagesStorage.String())
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
	return utils.CustomTagsMetadataList(ctx, SuiteData.StagesStorage)
}
