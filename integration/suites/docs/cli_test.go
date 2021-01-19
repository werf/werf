package docs_test

import (
	"runtime"

	. "github.com/onsi/ginkgo"
	"github.com/werf/werf/integration/pkg/utils"
)

var _ = Describe("docs", func() {
	BeforeEach(func() {
		if runtime.GOOS == "windows" {
			Skip("skip on windows")
		}

		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "../../../", "initial commit")

		SuiteData.Stubs.UnsetEnv("DOCKER_CONFIG")
		SuiteData.Stubs.UnsetEnv("WERF_DOCKER_CONFIG")
		SuiteData.Stubs.SetEnv("WERF_LOG_TERMINAL_WIDTH", "100")
	})

	It("should be without changes", func() {
		_, _ = utils.RunCommandWithOptions(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			[]string{"docs", "--dir", SuiteData.GetProjectWorktree(SuiteData.ProjectName)},
			utils.RunCommandOptions{ShouldSucceed: true, ExtraEnv: []string{"HOME=~", "WERF_PROJECT_NAME="}},
		)

		utils.RunSucceedCommand(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			"git",
			"add", "-A",
		)

		utils.RunSucceedCommand(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			"git",
			"diff", "--exit-code", "HEAD", "--",
		)
	})
})
