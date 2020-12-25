package ci_env_test

import (
	"testing"

	"github.com/werf/werf/integration/suite_init"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("CI-env suite", suite_init.TestSuiteEntrypointFuncOptions{})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData struct {
	suite_init.SuiteData
	TestDirPath string
}

var _ = SuiteData.StubsData.Setup()
var _ = SuiteData.SynchronizedSuiteCallbacksData.Setup()
var _ = SuiteData.WerfBinaryData.Setup(&SuiteData.SynchronizedSuiteCallbacksData)
