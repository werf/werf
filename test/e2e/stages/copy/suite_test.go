package e2e_stages_copy_test

import (
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/utils/docker"
)

func TestSuite(t *testing.T) {
	requiredTools := []string{"docker", "git"}
	if runtime.GOOS == "linux" {
		requiredTools = append(requiredTools, "buildah")
	}

	suite_init.MakeTestSuiteEntrypointFunc("E2E stages copy suite", suite_init.TestSuiteEntrypointFuncOptions{
		RequiredSuiteTools: requiredTools,
	})(t)
}

var SuiteData = struct {
	suite_init.SuiteData

	FromRegistryLocalAddress    string
	FromRegistryInternalAddress string
	FromRegistryContainerName   string

	ToRegistryLocalAddress    string
	ToRegistryInternalAddress string
	ToRegistryContainerName   string

	WerfFromAddr    string
	WerfToAddr      string
	WerfArchiveAddr string
}{}

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())

	_ = SuiteData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(_ []byte) {
		SuiteData.FromRegistryLocalAddress, SuiteData.FromRegistryInternalAddress, SuiteData.FromRegistryContainerName = docker.LocalDockerRegistryRun()
	})
	_ = SuiteData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(_ []byte) {
		SuiteData.ToRegistryLocalAddress, SuiteData.ToRegistryInternalAddress, SuiteData.ToRegistryContainerName = docker.LocalDockerRegistryRun()
	})

	_ = AfterEach(func(ctx SpecContext) {
		utils.RunSucceedCommand(ctx, "", SuiteData.WerfBinPath, "host", "purge", "--force", "--project-name", SuiteData.ProjectName)
	})
)
