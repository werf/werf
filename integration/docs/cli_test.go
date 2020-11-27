package docs_test

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/testing/utils"
)

var _ = Describe("docs", func() {
	BeforeEach(func() {
		if runtime.GOOS == "windows" {
			Skip("skip on windows")
		}

		resolvedExpectationPath, err := filepath.EvalSymlinks(utils.FixturePath("cli", "docs"))
		Ω(err).ShouldNot(HaveOccurred())

		utils.CopyIn(resolvedExpectationPath, filepath.Join(testDirPath, "docs"))

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"init",
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"add", "-A",
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"commit", "-m", "+",
		)

		stubs.UnsetEnv("DOCKER_CONFIG")
		stubs.UnsetEnv("WERF_DOCKER_CONFIG")
		stubs.SetEnv("WERF_LOG_TERMINAL_WIDTH", "100")
	})

	It("should be without changes", func() {
		_, _ = utils.RunCommandWithOptions(
			testDirPath,
			werfBinPath,
			[]string{"docs", "--dir", testDirPath},
			utils.RunCommandOptions{ShouldSucceed: true, ExtraEnv: []string{"HOME=~"}},
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"add", "-A",
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"diff-index", "--exit-code", "HEAD", "--",
		)
	})
})
