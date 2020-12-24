package docs_test

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/utils"
)

var _ = Describe("docs", func() {
	BeforeEach(func() {
		if runtime.GOOS == "windows" {
			Skip("skip on windows")
		}

		resolvedExpectationPath, err := filepath.EvalSymlinks(utils.FixturePath("cli", "docs"))
		Ω(err).ShouldNot(HaveOccurred())

		utils.CopyIn(resolvedExpectationPath, filepath.Join(SuiteData.TestDirPath, "docs"))

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"init",
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"add", "-A",
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"commit", "-m", "+",
		)

		SuiteData.Stubs.UnsetEnv("DOCKER_CONFIG")
		SuiteData.Stubs.UnsetEnv("WERF_DOCKER_CONFIG")
		SuiteData.Stubs.SetEnv("WERF_LOG_TERMINAL_WIDTH", "100")
	})

	It("should be without changes", func() {
		_, _ = utils.RunCommandWithOptions(
			SuiteData.TestDirPath,
			SuiteData.WerfBinPath,
			[]string{"docs", "--dir", SuiteData.TestDirPath},
			utils.RunCommandOptions{ShouldSucceed: true, ExtraEnv: []string{"HOME=~"}},
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"add", "-A",
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"diff-index", "--exit-code", "HEAD", "--",
		)
	})
})
