package helm_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/testing/utils"
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
			"helm", "delete", releaseName,
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
				"helm", "deploy-chart", ".", releaseName, "--namespace", releaseNamespace,
			)
		})

		When("first release has been deployed", func() {
			BeforeEach(func() {
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"helm", "deploy-chart", ".", releaseName, "--namespace", releaseNamespace,
				)
			})

			It("should get release templates and values", func() {
				output := utils.SucceedCommandOutputString(
					testDirPath,
					werfBinPath,
					"helm", "get", releaseName,
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
						"helm", "deploy-chart", ".", releaseName, "--namespace", releaseNamespace,
					)
				})

				It("should get release templates and values", func() {
					output := utils.SucceedCommandOutputString(
						testDirPath,
						werfBinPath,
						"helm", "get", releaseName,
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
						"helm", "list", releaseName,
					)

					Ω(output).Should(ContainSubstring(releaseName))
				})

				It("should get release history", func() {
					output := utils.SucceedCommandOutputString(
						testDirPath,
						werfBinPath,
						"helm", "history", releaseName,
					)

					Ω(strings.Count(output, "SUPERSEDED")).Should(BeEquivalentTo(1))
					Ω(strings.Count(output, "DEPLOYED")).Should(BeEquivalentTo(1))
				})

				It("should rollback release", func() {
					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"helm", "rollback", releaseName, "1",
					)

					output := utils.SucceedCommandOutputString(
						testDirPath,
						werfBinPath,
						"helm", "get", releaseName,
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
				"helm", "repo", "init",
			)
		})

		It("should deploy chart by chart reference", func() {
			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"helm", "deploy-chart", "stable/nginx-ingress", releaseName, "--namespace", releaseNamespace,
			)
		})
	})
})
