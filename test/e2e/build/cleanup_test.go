package e2e_build_test

import (
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/report"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.Describe("Test image cleanup", ginkgo.Label("e2e", "build", "extra", "docker"), func() {
	ginkgo.It("removes project registry tags while retaining another project's alias", func(ctx ginkgo.SpecContext) {
		setupEnv(setupEnvOptions{ContainerBackendMode: "docker"})
		SuiteData.InitTestRepo(ctx, "repo0", "simple/state0")
		project := report.NewProjectWithReport(newWerfProject("repo0"))
		_, buildReport := project.BuildWithReport(ctx, SuiteData.GetBuildReportPath("cleanup.json"), nil)
		imageRef := buildReport.Images["dockerfile"].DockerImageName
		registryRef := "localhost:65534/" + SuiteData.ProjectName + ":local-copy"
		foreignRef := "werf-test-neighbor-" + utils.GetRandomString(10) + ":keep"
		utils.RunSucceedCommand(ctx, "", "docker", "tag", imageRef, registryRef)
		utils.RunSucceedCommand(ctx, "", "docker", "tag", imageRef, foreignRef)
		ginkgo.DeferCleanup(func(cleanupCtx ginkgo.SpecContext) {
			utils.RunSucceedCommand(cleanupCtx, "", "docker", "rmi", "--no-prune", foreignRef)
		}, ginkgo.NodeTimeout(time.Minute))

		gomega.Expect(contback.CleanupProject(ctx, SuiteData.ProjectName, contback.CleanupProjectOptions{})).To(gomega.Succeed())
		for _, ref := range []string{imageRef, registryRef} {
			_, err := utils.RunCommand(ctx, "", "docker", "image", "inspect", ref)
			gomega.Expect(err).To(gomega.HaveOccurred())
		}
		utils.RunSucceedCommand(ctx, "", "docker", "image", "inspect", foreignRef)
	})
})
