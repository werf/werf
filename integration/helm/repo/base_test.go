// +build integration

package repo_test

import (
	"os"
	"path/filepath"

	"github.com/flant/werf/integration/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("helm repo", func() {
	BeforeEach(func() {
		Ω(os.Setenv("WERF_HELM_HOME", filepath.Join(testDirPath, ".helm"))).Should(Succeed())
	})

	It("helm should be configured", func() {
		output := utils.SucceedCommandOutput(
			testDirPath,
			werfBinPath,
			"helm", "repo", "init",
		)

		for _, substr := range []string{
			"Adding stable repo with URL",
			"Adding local repo with URL",
			"helm has been configured",
		} {
			Ω(output).Should(ContainSubstring(substr))
		}
	})

	Context("when chart repositories configuration is initialized", func() {
		BeforeEach(func() {
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"helm", "repo", "init",
			)
		})

		It("should update chart repositories cache", func() {
			output := utils.SucceedCommandOutput(
				testDirPath,
				werfBinPath,
				"helm", "repo", "update",
			)

			Ω(output).Should(ContainSubstring("Successfully got an update from the \"stable\" chart repository"))
		})

		It("default repositories should be listed", func() {
			output := utils.SucceedCommandOutput(
				testDirPath,
				werfBinPath,
				"helm", "repo", "list",
			)

			for _, substr := range []string{
				"stable",
				"local",
			} {
				Ω(output).Should(ContainSubstring(substr))
			}
		})

		It("should remove repository", func() {
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"helm", "repo", "remove", "stable",
			)

			output := utils.SucceedCommandOutput(
				testDirPath,
				werfBinPath,
				"helm", "repo", "list",
			)

			Ω(output).ShouldNot(ContainSubstring("stable"))
		})

		It("should add repository", func() {
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"helm", "repo", "add", "company", "https://kubernetes-charts.storage.googleapis.com",
			)

			output := utils.SucceedCommandOutput(
				testDirPath,
				werfBinPath,
				"helm", "repo", "list",
			)

			Ω(output).Should(ContainSubstring("company"))
		})

		It("should search charts", func() {
			output := utils.SucceedCommandOutput(
				testDirPath,
				werfBinPath,
				"helm", "repo", "search",
			)

			Ω(output).Should(ContainSubstring("stable/moodle"))
			Ω(output).ShouldNot(ContainSubstring("No results found"))
		})

		It("should fetch chart", func() {
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"helm", "repo", "fetch", "stable/moodle",
			)
		})
	})
})
