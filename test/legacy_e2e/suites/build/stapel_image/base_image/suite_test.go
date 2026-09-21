package base_image_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Ansible suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"docker"},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData suite_init.SuiteData

var _ = AfterEach(func(ctx SpecContext) {
	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "host", "purge", "--force")
})

var (
	_ = SuiteData.SetupStubs(suite_init.NewStubsData())
	_ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
	_ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
	_ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
	_ = SuiteData.SetupTmp(suite_init.NewTmpDirData())
)
