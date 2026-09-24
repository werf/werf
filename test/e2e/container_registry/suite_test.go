package e2e_container_registry_test

import (
	"testing"

	"github.com/werf/werf/v3/test/pkg/suite_init"
)

func TestSuite(t *testing.T) {
	suite_init.MakeTestSuiteEntrypointFunc("E2E container registry suite", suite_init.TestSuiteEntrypointFuncOptions{
		SuiteLabels: []string{suite_init.LabelNeedsRegistry},
	})(t)
}

var SuiteData struct {
	suite_init.SuiteData
}

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())
	_ = SuiteData.SetupWerfInit(suite_init.NewWerfInitData(SuiteData.TmpDirData))
	_ = SuiteData.SetupContainerRegistryPerImplementation(suite_init.NewContainerRegistryPerImplementationData(SuiteData.SynchronizedSuiteCallbacksData, true))
)
