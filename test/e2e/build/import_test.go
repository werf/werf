package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
)

var _ = Describe("Import", Label("e2e", "build", "import", "simple"), func() {
	DescribeTable("should resolve relative symlink destination",
		func(ctx SpecContext, testOpts setupEnvOptions) {
			By("initializing")
			setupEnv(testOpts)
			contRuntime := contback.NewContainerBackend(testOpts.ContainerBackendMode)

			By("building")
			repoDirname := "repo0"
			fixtureRelPath := "import/symlink_dest/state0"
			buildReportName := "report0.json"

			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			werfProject := newWerfProject(repoDirname)
			reportProject := report.NewProjectWithReport(werfProject)
			_, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)

			By("checking imported files landed in real dir and are accessible via symlink")
			contRuntime.ExpectCmdsToSucceed(
				ctx,
				buildReport.Images["target"].DockerImageName,
				"test -L /bin",
				"test -f /usr/bin/myapp",
				"echo 'hello' | diff /usr/bin/myapp -",
				"test -f /bin/myapp",
			)
		},
		backendEntry("Docker", setupEnvOptions{
			ContainerBackendMode:        "docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
		backendEntry("Native Buildah rootless", setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
	)
})
