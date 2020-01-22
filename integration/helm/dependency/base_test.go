package dependency_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
)

var _ = Describe("helm dependency", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("default"), testDirPath)

		stubs.SetEnv("WERF_HELM_HOME", filepath.Join(testDirPath, ".helm"))
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"helm", "repo", "init",
		)
	})

	It("should be listed", func() {
		output := utils.SucceedCommandOutputString(
			testDirPath,
			werfBinPath,
			"helm", "dependency", "list",
		)

		for _, substr := range []string{
			"mysql",
			"redis",
			"rabbitmq",
		} {
			Ω(output).Should(ContainSubstring(substr))
		}
	})

	It("should be built", func() {
		output := utils.SucceedCommandOutputString(
			testDirPath,
			werfBinPath,
			"helm", "dependency", "build",
		)

		for _, substr := range []string{
			"Downloading mysql from repo",
			"Downloading redis from repo",
			"Downloading rabbitmq from repo",
		} {
			Ω(output).Should(ContainSubstring(substr))
		}
	})

	It("should be updated", func() {
		output := utils.SucceedCommandOutputString(
			testDirPath,
			werfBinPath,
			"helm", "dependency", "build",
		)

		for _, substr := range []string{
			"Downloading mysql from repo",
			"Downloading redis from repo",
			"Downloading rabbitmq from repo",
		} {
			Ω(output).Should(ContainSubstring(substr))
		}
	})
})
