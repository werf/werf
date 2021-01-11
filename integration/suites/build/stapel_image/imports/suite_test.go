package imports_test

import (
	"testing"

	"github.com/werf/werf/integration/pkg/suite_init"

	"github.com/onsi/ginkgo"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Imports suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"docker", "git"},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData suite_init.SuiteData

var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
var _ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
var _ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))

var _ = ginkgo.BeforeEach(func() {
	SuiteData.Stubs.SetEnv("WERF_DISABLE_AUTO_GC", "1")
})
