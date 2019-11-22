// +build integration

package dependency_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/integration/utils"
)

var _ = Describe("helm dependency", func() {
	BeforeEach(func() {
		utils.CopyIn(fixturePath("default"), testDirPath)

		Ω(os.Setenv("WERF_HELM_HOME", filepath.Join(testDirPath, ".helm"))).Should(Succeed())
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
