package cleanup_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/test/pkg/suite_init"
	"github.com/werf/werf/test/pkg/utils"
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

func perImplementationBeforeEach(implementationName string) func() {
	return func() {
		werfImplementationName := SuiteData.ContainerRegistryPerImplementation[implementationName].WerfImplementationName

		repo := fmt.Sprintf("%s/%s", SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryAddress, SuiteData.ProjectName)
		InitStagesStorage(repo, werfImplementationName, SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryOptions)
		SuiteData.SetupRepo(context.Background(), repo, implementationName, SuiteData.StubsData)
		SuiteData.TestImplementation = implementationName

		containerRegistry, err := docker_registry.NewDockerRegistry(repo, werfImplementationName, docker_registry.DockerRegistryOptions{})
		Ω(err).ShouldNot(HaveOccurred())

		SuiteData.ContainerRegistry = containerRegistry
	}
}

func InitStagesStorage(stagesStorageAddress, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) {
	SuiteData.StagesStorage = utils.NewStagesStorage(stagesStorageAddress, implementationName, dockerRegistryOptions)
}

func StagesCount() int {
	return utils.StagesCount(context.Background(), SuiteData.StagesStorage)
}

func ManagedImagesCount() int {
	return utils.ManagedImagesCount(context.Background(), SuiteData.StagesStorage)
}

func ImageMetadata(imageName string) map[string][]string {
	return utils.ImageMetadata(context.Background(), SuiteData.StagesStorage, imageName)
}

func RmImportMetadata(importSourceID string) {
	utils.RmImportMetadata(context.Background(), SuiteData.StagesStorage, importSourceID)
}

func ImportMetadataIDs() []string {
	return utils.ImportMetadataIDs(context.Background(), SuiteData.StagesStorage)
}

func CustomTags() []string {
	tags, err := SuiteData.ContainerRegistry.Tags(context.Background(), SuiteData.StagesStorage.String())
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
