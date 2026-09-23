package e2e_helm_secret_test

import (
	"testing"

	"github.com/werf/werf/v2/test/pkg/suite_init"
)

func TestSuite(t *testing.T) {
	suite_init.MakeTestSuiteEntrypointFunc("E2E helm secret suite", suite_init.TestSuiteEntrypointFuncOptions{
		RequiredSuiteTools: []string{"git"},
	})(t)
}

var SuiteData struct {
	suite_init.SuiteData
}

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())
)
