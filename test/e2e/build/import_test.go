package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type importTestOptions struct {
	setupEnvOptions
}

var _ = Describe("Import", Label("e2e", "build", "import", "simple"), func() {
	DescribeTable("should resolve relative symlink destination",
		func(ctx SpecContext, testOpts importTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("building")
			repoDirname := "repo0"
			fixtureRelPath := "import/symlink_dest/state0"
			buildReportName := "report0.json"

			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
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
		Entry("Vanilla Docker", importTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("BuildKit Docker", importTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("Native Buildah rootless", importTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("Native Buildah chroot", importTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: true,
		}}),
		Entry("Native Buildah rootless", importTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("Native Buildah chroot", importTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: true,
		}}),
	)
})
