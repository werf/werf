package imports_test

import (
	"testing"

	"github.com/werf/werf/integration/suite_init"

	"github.com/onsi/ginkgo"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Imports suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"docker", "git"},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData suite_init.SuiteData

var _ = SuiteData.StubsData.Setup()
var _ = SuiteData.SynchronizedSuiteCallbacksData.Setup()
var _ = SuiteData.WerfBinaryData.Setup(&SuiteData.SynchronizedSuiteCallbacksData)
var _ = SuiteData.ProjectNameData.Setup(&SuiteData.StubsData)

var _ = ginkgo.BeforeEach(func() {
	SuiteData.Stubs.SetEnv("WERF_DISABLE_AUTO_GC", "1")
})
