package git_test

import (
	"testing"

	"github.com/onsi/ginkgo"

	"github.com/werf/werf/integration/suite_init"

	"github.com/werf/werf/integration/utils"
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
		"purge", "--force",
	)
})

var _ = SuiteData.StubsData.Setup()
var _ = SuiteData.SynchronizedSuiteCallbacksData.Setup()
var _ = SuiteData.WerfBinaryData.Setup(&SuiteData.SynchronizedSuiteCallbacksData)
var _ = SuiteData.ProjectNameData.Setup(&SuiteData.StubsData)
var _ = SuiteData.TmpDirData.Setup()

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
