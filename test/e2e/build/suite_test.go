package e2e_build_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v3/test/pkg/suite_init"
	"github.com/werf/werf/v3/test/pkg/utils"
)

func TestSuite(t *testing.T) {
	suite_init.MakeTestSuiteEntrypointFunc("E2E Build suite", suite_init.TestSuiteEntrypointFuncOptions{
		RequiredSuiteTools: []string{"docker", "git"},
	})(t)
}

var SuiteData = struct {
	suite_init.SuiteData

	WerfRepo string

	TempFiles []string
}{}

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())

	_ = SuiteData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(ctx context.Context, _ []byte) {
		SuiteData.TempFiles = append([]string{}, utils.CreateTmpFileInHome("secret_file_in_home", "secret"))
	})

	_ = AfterEach(func(ctx SpecContext) {
		utils.RunSucceedCommand(ctx, "", SuiteData.WerfBinPath, "host", "purge", "--force", "--project-name", SuiteData.ProjectName)
	})
)
