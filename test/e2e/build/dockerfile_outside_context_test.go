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
	DescribeTable("should build image with the build context assembled by werf only",
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

			By("building images")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath("repo0"))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report0.json"), nil)
			Expect(buildOut).To(ContainSubstring("Building stage"))
			Expect(buildOut).NotTo(ContainSubstring("There is no way to ignore the Dockerfile"))

			By("checking that the dockerfile-specific .dockerignore wins over every other ignore file")
			contRuntime.ExpectCmdsToSucceed(
				ctx,
				buildReport.Images["dockerfile"].DockerImageName,
				// the dockerfile-specific .containerignore matches this file
				"echo 'filecontent' | diff /ctx/file -",
				// the ignore files of the context match this one, and only the container
				// backend could have dropped it
				"echo 'keep' | diff /ctx/keepme -",
				"echo 'generated' | diff /ctx/generated.txt -",
				"test ! -e /ctx/ignored",
				"test 5 = $(ls -A /ctx | wc -l)",
			)

			By("checking that .dockerignore of the context wins over .containerignore of the context")
			contRuntime.ExpectCmdsToSucceed(
				ctx,
				buildReport.Images["dockerignore-priority"].DockerImageName,
				"echo 'keep' | diff /ctx/keep -",
				"echo 'only-ci' | diff /ctx/only-ci -",
				"test ! -e /ctx/only-di",
				"test 4 = $(ls -A /ctx | wc -l)",
			)

			By("checking that .containerignore of the context is applied when there is nothing else")
			contRuntime.ExpectCmdsToSucceed(
				ctx,
				buildReport.Images["containerignore"].DockerImageName,
				"echo 'filecontent' | diff /ctx/file -",
				"test ! -e /ctx/ignored",
				"test 2 = $(ls -A /ctx | wc -l)",
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
