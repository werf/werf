package docs_test

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
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
			utils.RunCommandOptions{ShouldSucceed: true, ExtraEnv: []string{"HOME=~", "WERF_PROJECT_NAME="}},
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"add", "-A",
		)

		// Exclude docs/_includes/reference/cli/werf_kubectl_get.md because options order of the command is not persistent, and we do not manage kubectl commands.
		// -werf kubectl get [(-o|--output=)json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-as-json|jsonpath-file|custom-columns-file|custom-columns|wide] (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...) [flags] [options]
		// +werf kubectl get [(-o|--output=)json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-as-json|jsonpath-file|custom-columns|custom-columns-file|wide] (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...) [flags] [options]
		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"diff", "--exit-code", "HEAD", "--", ":(exclude)docs/_includes/reference/cli/werf_kubectl_get.md", ".",
		)
	})
})
