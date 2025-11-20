package ci_env_test

import (
	"testing"

	"github.com/werf/werf/v2/test/pkg/suite_init"
)

func TestSuite(t *testing.T) {
	requiredTools := []string{"docker", "git"}
	suite_init.MakeTestSuiteEntrypointFunc("E2E ci-env suite", suite_init.TestSuiteEntrypointFuncOptions{
		RequiredSuiteTools: requiredTools,
	})(t)
}

var SuiteData = struct {
	suite_init.SuiteData
}{}

var (
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())
)
