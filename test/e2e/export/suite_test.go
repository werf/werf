package e2e_export_test

import (
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
)

func TestExport(t *testing.T) {
	requiredTools := []string{"docker", "git"}
	if runtime.GOOS == "linux" {
		requiredTools = append(requiredTools, "buildah")
	}
	suite_init.MakeTestSuiteEntrypointFunc("E2E Export suite", suite_init.TestSuiteEntrypointFuncOptions{
		RequiredSuiteTools: requiredTools,
		RequiredSuiteEnvs: []string{
			"WERF_TEST_K8S_DOCKER_REGISTRY",
		},
	})(t)
}

var SuiteData = struct {
	suite_init.SuiteData
}{}

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())

	_ = AfterEach(func(ctx SpecContext) {
		utils.RunSucceedCommand(ctx, "", SuiteData.WerfBinPath, "host", "purge", "--force", "--project-name", SuiteData.ProjectName)
	})
)
