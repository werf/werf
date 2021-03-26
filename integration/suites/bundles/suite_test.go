package bundles_test

import (
	"testing"

	"github.com/werf/werf/integration/pkg/suite_init"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Bundles suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"git", "docker"},
	RequiredSuiteEnvs:  []string{},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData struct {
	suite_init.SuiteData
	Repo string
}

var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
var _ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
var _ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
var _ = SuiteData.SetupContainerRegistryPerImplementation(suite_init.NewContainerRegistryPerImplementationData(SuiteData.SynchronizedSuiteCallbacksData, false))
