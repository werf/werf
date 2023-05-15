package deploy_test

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

const customTagValue = "custom-tag"

var _ = Describe("custom tag", func() {
	BeforeEach(func() {
		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "custom_tag", "initial commit")
		SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")
	})

	When("should do release", func() {
		_ = AfterEach(func() {
			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"dismiss", "--with-namespace",
			)

			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"purge",
			)
		})

		It("should use custom tag", func() {
			var werfArgs []string
			werfArgs = append(werfArgs, "converge")
			werfArgs = append(
				werfArgs,
				"--set", fmt.Sprintf("imageCredentials.registry=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY")),
				"--set", fmt.Sprintf("imageCredentials.username=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME")),
				"--set", fmt.Sprintf("imageCredentials.password=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD")),
			)
			werfArgs = append(werfArgs, "--use-custom-tag", customTagValue)

			utils.RunSucceedCommand(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				werfArgs...,
			)

			output := utils.SucceedCommandOutputString(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"helm", "get-release",
			)
			releaseName := strings.TrimSpace(output)

			output = utils.SucceedCommandOutputString(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"helm", "get-namespace",
			)
			namespaceName := strings.TrimSpace(output)

			output = utils.SucceedCommandOutputString(
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"helm", "get", "manifest", "--namespace", namespaceName, releaseName,
			)

			repoCustomTag := strings.Join([]string{SuiteData.K8sDockerRegistryRepo, customTagValue}, ":")
			expectedSubstring := "image: " + repoCustomTag
			Ω(output).Should(ContainSubstring(expectedSubstring))
		})
	})

	It("should fail with outdated custom tag", func() {
		SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")

		utils.RunSucceedCommand(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			"build", "--add-custom-tag", customTagValue,
		)

		SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "2")

		utils.RunSucceedCommand(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			"build",
		)

		var werfArgs []string
		werfArgs = append(werfArgs, "converge")
		werfArgs = append(
			werfArgs,
			"--set", fmt.Sprintf("imageCredentials.registry=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY")),
			"--set", fmt.Sprintf("imageCredentials.username=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME")),
			"--set", fmt.Sprintf("imageCredentials.password=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD")),
		)
		werfArgs = append(werfArgs, "--use-custom-tag", customTagValue)
		werfArgs = append(werfArgs, "--require-built-images")

		bytes, err := utils.RunCommand(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			werfArgs...,
		)

		Ω(err).Should(HaveOccurred())
		Ω(string(bytes)).Should(ContainSubstring("custom tag \"custom-tag\" image must be the same as associated content-based tag"))
	})
})
