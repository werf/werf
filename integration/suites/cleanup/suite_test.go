package cleanup_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"

	"github.com/werf/werf/integration/pkg/suite_init"
	"github.com/werf/werf/integration/pkg/utils"
)

const imageName = "image"

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Cleanup suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"git", "docker"},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData struct {
	suite_init.SuiteData
	TestImplementation string
	StagesStorage      storage.StagesStorage
}

var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
var _ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
var _ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
var _ = SuiteData.SetupTmp(suite_init.NewTmpDirData())
var _ = SuiteData.SetupContainerRegistryPerImplementation(suite_init.NewContainerRegistryPerImplementationData(SuiteData.SynchronizedSuiteCallbacksData, true))

func perImplementationBeforeEach(implementationName string) func() {
	return func() {
		werfImplementationName := SuiteData.ContainerRegistryPerImplementation[implementationName].WerfImplementationName

		repo := fmt.Sprintf("%s/%s", SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryAddress, SuiteData.ProjectName)
		InitStagesStorage(repo, werfImplementationName, SuiteData.ContainerRegistryPerImplementation[implementationName].RegistryOptions)
		SuiteData.SetupRepo(context.Background(), repo, implementationName, SuiteData.StubsData)
		SuiteData.TestImplementation = implementationName
	}
}

func InitStagesStorage(stagesStorageAddress string, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) {
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
