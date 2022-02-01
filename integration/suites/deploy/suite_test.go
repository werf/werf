package deploy_test

import (
	"testing"

	"github.com/werf/werf/test/pkg/suite_init"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Deploy suite", suite_init.TestSuiteEntrypointFuncOptions{})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData suite_init.SuiteData

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupK8sDockerRegistry(suite_init.NewK8sDockerRegistryData(SuiteData.ProjectNameData, SuiteData.StubsData))
)
