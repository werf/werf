package deploy_test

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

const customTagValue = "custom-tag"

var _ = Describe("custom tag", func() {
	BeforeEach(func(ctx SpecContext) {
		SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, "custom_tag", "initial commit")
		SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")
	})

	When("should do release", func() {
		_ = AfterEach(func(ctx SpecContext) {
			utils.RunSucceedCommand(
				ctx,
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"dismiss", "--with-namespace",
			)

			utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "purge")
		})

		It("should use custom tag", func(ctx SpecContext) {
			var werfArgs []string
			werfArgs = append(werfArgs, "converge")
			werfArgs = append(
				werfArgs,
				"--set", fmt.Sprintf("imageCredentials.registry=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY")),
				"--set", fmt.Sprintf("imageCredentials.username=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME")),
				"--set", fmt.Sprintf("imageCredentials.password=%s", os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD")),
			)
			werfArgs = append(werfArgs, "--use-custom-tag", customTagValue)

			utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, werfArgs...)

			output := utils.SucceedCommandOutputString(
				ctx,
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"helm", "get-release",
			)
			releaseName := strings.TrimSpace(output)

			output = utils.SucceedCommandOutputString(
				ctx,
				SuiteData.GetProjectWorktree(SuiteData.ProjectName),
				SuiteData.WerfBinPath,
				"helm", "get-namespace",
			)
			namespaceName := strings.TrimSpace(output)

			output = utils.SucceedCommandOutputString(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "helm", "get", "manifest", "--namespace", namespaceName, releaseName)

			repoCustomTag := strings.Join([]string{SuiteData.K8sDockerRegistryRepo, customTagValue}, ":")
			expectedSubstring := "image: " + repoCustomTag
			Expect(output).Should(ContainSubstring(expectedSubstring))
		})
	})

	It("should fail with outdated custom tag", func(ctx SpecContext) {
		SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "1")

		utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "build", "--add-custom-tag", customTagValue)

		SuiteData.Stubs.SetEnv("FROM_CACHE_VERSION", "2")

		utils.RunSucceedCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, "build")

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

		bytes, err := utils.RunCommand(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, werfArgs...)

		Expect(err).Should(HaveOccurred())
		Expect(string(bytes)).Should(ContainSubstring("custom tag \"custom-tag\" image must be the same as associated content-based tag"))
	})
})
