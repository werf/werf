package e2e_export_test

import (
	"testing"

	"github.com/werf/werf/v3/test/pkg/suite_init"
)

func TestExport(t *testing.T) {
	suite_init.MakeTestSuiteEntrypointFunc("E2E Export suite", suite_init.TestSuiteEntrypointFuncOptions{
		RequiredSuiteTools: []string{"docker", "git"},
		SuiteLabels:        []string{suite_init.LabelNeedsRegistry},
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

	_ = SuiteData.SetupProjectCleanup()
)
