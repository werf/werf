package cleanup_test

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = Describe("host purge command", func() {
	SuiteData.SetupProjectCleanup()

	BeforeEach(func(ctx SpecContext) {
		Expect(werf.Init(SuiteData.TmpDir, "")).To(Succeed())
		SuiteData.Stubs.UnsetEnv("WERF_REPO")

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
			var id string
			containerExists := func(ctx SpecContext) bool {
				return utils.SucceedCommandOutputString(ctx, "", "docker", "ps", "--all", "--quiet", "--filter", "id="+id) != ""
			}
			BeforeEach(func(ctx SpecContext) {
				utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "build")

				const cidFile = "purge-container.cid"
				out, runErr := utils.RunCommand(ctx, SuiteData.TestDirPath, SuiteData.WerfBinPath, "run", "--docker-options", "-d --cidfile="+cidFile, "--", "/bin/sleep", "3600")
				content, readErr := os.ReadFile(filepath.Join(SuiteData.TestDirPath, cidFile))
				id = strings.TrimSpace(string(content))
				if id != "" {
					DeferCleanup(func(ctx SpecContext) {
						if containerExists(ctx) {
							utils.RunSucceedCommand(ctx, "", "docker", "rm", "--force", id)
						}
					}, NodeTimeout(time.Minute))
				}
				Expect(runErr).NotTo(HaveOccurred(), string(out))
				Expect(readErr).NotTo(HaveOccurred())
				Expect(id).NotTo(BeEmpty())
				Expect(containerExists(ctx)).To(BeTrue())
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
				Expect(containerExists(ctx)).To(BeFalse())
			})
		})
	})
})
