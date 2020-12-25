package secret_test

import (
	"testing"

	"github.com/werf/werf/integration/suite_init"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm/Get Suite")
}

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Helm/Secret suite", suite_init.TestSuiteEntrypointFuncOptions{})

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
var _ = SuiteData.ProjectNameData.Setup(&SuiteData.StubsData)
var _ = SuiteData.TmpDirData.Setup()
