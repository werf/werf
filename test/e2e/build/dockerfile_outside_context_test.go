package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type dockerfileOutsideContextTestOptions struct {
	setupEnvOptions
}

var _ = Describe("Build with dockerfile outside context", Label("e2e", "build", "dockerfile-outside-context"), func() {
	DescribeTable("should build image and keep the dockerfile out of the build context",
		func(ctx SpecContext, testOpts dockerfileOutsideContextTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("preparing test repo")
			SuiteData.InitTestRepo(ctx, "repo0", "dockerfile_outside_context/state0")

			By("building image")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath("repo0"))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report0.json"), nil)
			Expect(buildOut).To(ContainSubstring("Building stage"))

			By("checking build context contents")
			contRuntime.ExpectCmdsToSucceed(
				ctx,
				buildReport.Images["dockerfile"].DockerImageName,
				"echo 'filecontent' | diff /ctx/file -",
				"test 1 = $(ls -A /ctx | wc -l)",
			)
		},
		Entry("using Docker", dockerfileOutsideContextTestOptions{setupEnvOptions{
			ContainerBackendMode:        "docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("using Native Buildah with chroot isolation", dockerfileOutsideContextTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})
