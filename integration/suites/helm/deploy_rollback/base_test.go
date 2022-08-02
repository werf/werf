package deploy_rollback_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("deploy and rollback chart", func() {
	var releaseName string
	var releaseNamespace string

	BeforeEach(func() {
		releaseName = utils.ProjectName()
		releaseNamespace = releaseName
	})

	When("deploy local chart", func() {
		AfterEach(func() {
			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"helm", "uninstall", releaseName, "--namespace", releaseNamespace,
			)
		})

		BeforeEach(func() {
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("chart_1"), "initial commit")
		})

		It("should deploy chart in working directory", func() {
			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"helm", "install", releaseName, ".", "--namespace", releaseNamespace,
			)
		})

		When("first release has been deployed", func() {
			BeforeEach(func() {
				utils.RunSucceedCommand(
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
					SuiteData.WerfBinPath,
					"helm", "install", releaseName, ".", "--namespace", releaseNamespace,
				)
			})

			It("should get release templates and values", func() {
				output := utils.SucceedCommandOutputString(
					SuiteData.GetProjectWorktree(SuiteData.ProjectName),
					SuiteData.WerfBinPath,
					"helm", "get", "all", releaseName, "--namespace", releaseNamespace,
				)

				expectedSubStrings := []string{
					"REVISION: 1",
					"COMPUTED VALUES",
					"HOOKS",
					"MANIFEST",
					"# Source: chart/templates/serviceaccount.yaml",
					"# Source: chart/templates/service.yaml",
					"# Source: chart/templates/deployment.yaml",
				}

				for _, expectedSubString := range expectedSubStrings {
					Ω(output).Should(ContainSubstring(expectedSubString))
				}
			})

			When("second release has been deployed", func() {
				BeforeEach(func() {
					SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("chart_2"), "initial commit")

					utils.RunSucceedCommand(
						SuiteData.GetProjectWorktree(SuiteData.ProjectName),
						SuiteData.WerfBinPath,
						"helm", "upgrade", releaseName, ".", "--namespace", releaseNamespace,
					)
				})

				It("should get release templates and values", func() {
					output := utils.SucceedCommandOutputString(
						SuiteData.GetProjectWorktree(SuiteData.ProjectName),
						SuiteData.WerfBinPath,
						"helm", "get", "all", releaseName, "--namespace", releaseNamespace,
					)

					expectedSubStrings := []string{
						"REVISION: 2",
						"COMPUTED VALUES",
						"HOOKS",
						"MANIFEST",
						"# Source: chart/templates/service.yaml",
					}

					notExpectedSubStrings := []string{
						"# Source: chart/templates/serviceaccount.yaml",
						"# Source: chart/templates/deployment.yaml",
					}

					for _, expectedSubString := range expectedSubStrings {
						Ω(output).Should(ContainSubstring(expectedSubString))
					}

					for _, notExpectedSubString := range notExpectedSubStrings {
						Ω(output).ShouldNot(ContainSubstring(notExpectedSubString))
					}
				})

				It("should list release", func() {
					output := utils.SucceedCommandOutputString(
						SuiteData.GetProjectWorktree(SuiteData.ProjectName),
						SuiteData.WerfBinPath,
						"helm", "list", "--namespace", releaseNamespace,
					)

					Ω(output).Should(ContainSubstring(releaseName))
				})

				It("should get release history", func() {
					output := utils.SucceedCommandOutputString(
						SuiteData.GetProjectWorktree(SuiteData.ProjectName),
						SuiteData.WerfBinPath,
						"helm", "history", releaseName, "--namespace", releaseNamespace,
					)

					Ω(strings.Count(output, "superseded")).Should(BeEquivalentTo(1))
					Ω(strings.Count(output, "deployed")).Should(BeEquivalentTo(1))
				})

				It("should rollback release", func() {
					utils.RunSucceedCommand(
						SuiteData.GetProjectWorktree(SuiteData.ProjectName),
						SuiteData.WerfBinPath,
						"helm", "rollback", releaseName, "1", "--namespace", releaseNamespace,
					)

					output := utils.SucceedCommandOutputString(
						SuiteData.GetProjectWorktree(SuiteData.ProjectName),
						SuiteData.WerfBinPath,
						"helm", "get", "all", releaseName, "--namespace", releaseNamespace,
					)

					Ω(output).Should(ContainSubstring("REVISION: 3"))
				})
			})
		})
	})

	When("deploy by chart reference", func() {
		AfterEach(func() {
			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"helm", "uninstall", releaseName, "--namespace", releaseNamespace,
			)
		})

		BeforeEach(func() {
			SuiteData.Stubs.SetEnv("XDG_DATA_HOME", SuiteData.TestDirPath)
			SuiteData.Stubs.SetEnv("XDG_CACHE_HOME", SuiteData.TestDirPath)
			SuiteData.Stubs.SetEnv("XDG_CONFIG_HOME", SuiteData.TestDirPath)

			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"helm", "repo", "add", "stable", "https://charts.helm.sh/stable",
			)
		})

		It("should deploy chart by chart reference", func() {
			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"helm", "install", releaseName, "stable/nginx-ingress", "--namespace", releaseNamespace,
			)
		})
	})
})
