package helm_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/utils"
)

var _ = Describe("deploy and rollback chart", func() {
	var releaseName string
	var releaseNamespace string

	BeforeEach(func() {
		releaseName = utils.ProjectName()
		releaseNamespace = releaseName
	})

	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"helm", "uninstall", releaseName, "--namespace", releaseNamespace,
		)
	})

	When("deploy local chart", func() {
		BeforeEach(func() {
			utils.CopyIn(utils.FixturePath("chart_1"), testDirPath)
		})

		It("should deploy chart in working directory", func() {
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"helm", "install", releaseName, ".", "--namespace", releaseNamespace,
			)
		})

		When("first release has been deployed", func() {
			BeforeEach(func() {
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"helm", "install", releaseName, ".", "--namespace", releaseNamespace,
				)
			})

			It("should get release templates and values", func() {
				output := utils.SucceedCommandOutputString(
					testDirPath,
					werfBinPath,
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
					Ω(os.RemoveAll(testDirPath)).ShouldNot(HaveOccurred())
					utils.CopyIn(utils.FixturePath("chart_2"), testDirPath)

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"helm", "upgrade", releaseName, ".", "--namespace", releaseNamespace,
					)
				})

				It("should get release templates and values", func() {
					output := utils.SucceedCommandOutputString(
						testDirPath,
						werfBinPath,
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
						testDirPath,
						werfBinPath,
						"helm", "list", "--namespace", releaseNamespace,
					)

					Ω(output).Should(ContainSubstring(releaseName))
				})

				It("should get release history", func() {
					output := utils.SucceedCommandOutputString(
						testDirPath,
						werfBinPath,
						"helm", "history", releaseName, "--namespace", releaseNamespace,
					)

					Ω(strings.Count(output, "superseded")).Should(BeEquivalentTo(1))
					Ω(strings.Count(output, "deployed")).Should(BeEquivalentTo(1))
				})

				It("should rollback release", func() {
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"helm", "rollback", releaseName, "1", "--namespace", releaseNamespace,
					)

					output := utils.SucceedCommandOutputString(
						testDirPath,
						werfBinPath,
						"helm", "get", "all", releaseName, "--namespace", releaseNamespace,
					)

					Ω(output).Should(ContainSubstring("REVISION: 3"))
				})
			})
		})
	})

	When("deploy by chart reference", func() {
		BeforeEach(func() {
			stubs.SetEnv("WERF_HELM_HOME", testDirPath)

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"helm", "repo", "add", "stable", "https://charts.helm.sh/stable",
			)
		})

		It("should deploy chart by chart reference", func() {
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"helm", "install", releaseName, "stable/nginx-ingress", "--namespace", releaseNamespace,
			)
		})
	})
})
