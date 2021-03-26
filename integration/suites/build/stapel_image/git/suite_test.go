package git_test

import (
	"testing"

	"github.com/onsi/ginkgo"

	"github.com/werf/werf/integration/pkg/suite_init"

	"github.com/werf/werf/integration/pkg/utils"
)

const gitCacheSizeStep = 1024 * 1024

var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("Build/Stapel Image/Git suite", suite_init.TestSuiteEntrypointFuncOptions{
	RequiredSuiteTools: []string{"docker", "git"},
})

func TestSuite(t *testing.T) {
	testSuiteEntrypointFunc(t)
}

var SuiteData suite_init.SuiteData

var _ = ginkgo.AfterEach(func() {
	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		SuiteData.WerfBinPath,
		"host", "purge", "--force",
	)
})

var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
var _ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
var _ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
var _ = SuiteData.SetupTmp(suite_init.NewTmpDirData())

func commonBeforeEach(fixturePath string) {
	utils.CopyIn(fixturePath, SuiteData.TestDirPath)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"init",
	)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"add", "werf*.yaml",
	)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"commit", "-m", "Initial commit",
	)
}
