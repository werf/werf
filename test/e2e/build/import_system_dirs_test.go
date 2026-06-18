package e2e_build_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Import system dirs", Label("e2e", "build", "import", "system-dirs"), func() {
	DescribeTable("should not import system directories when using wildcard includePaths",
		func(ctx SpecContext, testOpts setupEnvOptions) {
			By("initializing")
			setupEnv(testOpts)

			By("building")
			repoDirname := "repo0"
			fixtureRelPath := "import/no_system_dirs/state0"
			buildReportName := "report0.json"

			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			_, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)

			By("checking system directories are not present in destination image")
			imageName := buildReport.Images["destination"].DockerImageName
			checkNoSystemDirsInImage(ctx, imageName)
		},
		Entry("Vanilla Docker", setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
	)
})

func checkNoSystemDirsInImage(ctx context.Context, imageName string) {
	for _, dir := range []string{"proc", "sys", "dev", "run"} {
		out, err := utils.RunCommand(ctx, "/", "docker", "run", "--rm", "--entrypoint", "sh", imageName, "-c",
			"test ! -e /"+dir)
		Expect(err).NotTo(HaveOccurred(), "system dir /%s must not exist in image, output: %s", dir, string(out))
	}
}
