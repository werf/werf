package cleanup_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("host purge command", func() {
	BeforeEach(func(ctx SpecContext) {
		SuiteData.StagesStorage = utils.NewStagesStorage(ctx, ":local", "default", docker_registry.DockerRegistryOptions{})

		utils.CopyIn(utils.FixturePath("host_purge"), SuiteData.TestDirPath)

		utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "init")

		utils.RunSucceedCommand(
			ctx,
			SuiteData.TestDirPath,
			"git",
			"add", "werf.yaml",
		)

		utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "-m", "Initial commit")
	})

	When("project name specified", func() {
		Context("when there is running container based on werf image", func() {
			BeforeEach(func(ctx SpecContext) {
				utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

				utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "run", "--docker-options", "-d", "--", "/bin/sleep", "30")
			})

			It("should fail with specific error", func(ctx SpecContext) {
				out, err := utils.RunCommand(
					ctx,
					SuiteData.TestDirPath,
					SuiteData.WerfBinPath,
					"host", "purge",
				)
				Expect(err).ShouldNot(Succeed())
				Expect(string(out)).Should(ContainSubstring("used by container"))
				Expect(string(out)).Should(ContainSubstring("Use --force option to remove all containers that are based on deleting werf docker images"))
			})

			It("should remove project images and related container", func(ctx SpecContext) {
				utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "host", "purge", "--force")

				Expect(StagesCount(ctx)).Should(Equal(0))
			})
		})
	})
})
