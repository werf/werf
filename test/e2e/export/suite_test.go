package e2e_export_test

import (
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/utils/docker"
)

func TestExport(t *testing.T) {
	requiredTools := []string{"docker", "git"}
	if runtime.GOOS == "linux" {
		requiredTools = append(requiredTools, "buildah")
	}
	suite_init.MakeTestSuiteEntrypointFunc("E2E Export suite", suite_init.TestSuiteEntrypointFuncOptions{
		RequiredSuiteTools: requiredTools,
	})(t)
}

var SuiteData = struct {
	suite_init.SuiteData

	RegistryLocalAddress    string
	RegistryInternalAddress string
	RegistryContainerName   string
	ContainerRegistry       docker_registry.Interface

	WerfRepo string
}{}

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())
	_ = SuiteData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(_ []byte) {
		SuiteData.RegistryLocalAddress, SuiteData.RegistryInternalAddress, SuiteData.RegistryContainerName = docker.LocalDockerRegistryRun()
	})
	_ = SuiteData.AppendSynchronizedAfterSuiteAllNodesFunc(func() {
		docker.ContainerStopAndRemove(SuiteData.RegistryContainerName)
	})

	_ = AfterEach(func() {
		utils.RunSucceedCommand("", SuiteData.WerfBinPath, "host", "purge", "--force", "--project-name", SuiteData.ProjectName)
	})
)

var _ = BeforeEach(func() {
	containerRegistry, err := docker_registry.NewDockerRegistry(SuiteData.RegistryLocalAddress+"/repo", "", docker_registry.DockerRegistryOptions{
		InsecureRegistry:      true,
		SkipTlsVerifyRegistry: true,
	})
	Expect(err).ShouldNot(HaveOccurred())

	SuiteData.ContainerRegistry = containerRegistry
})
