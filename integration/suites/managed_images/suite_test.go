package managed_images_test

import (
	"testing"

	"github.com/werf/werf/integration/pkg/suite_init"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Managed Images suite", suite_init.TestSuiteEntrypointFuncOptions{})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData struct {
	suite_init.SuiteData
}

var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
var _ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
var _ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
var _ = SuiteData.SetupTmp(suite_init.NewTmpDirData())
var _ = SuiteData.SetupContainerRegistryPerImplementation(suite_init.NewContainerRegistryPerImplementationData(SuiteData.SynchronizedSuiteCallbacksData, true))
