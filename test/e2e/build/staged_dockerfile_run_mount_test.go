package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/suite_init"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type stagedDockerfileRunMountTestOptions struct {
	setupEnvOptions
}

var _ = Describe("Staged Dockerfile build with RUN --mount from stage", Label("e2e", "build", "staged_dockerfile_run_mount"), func() {
	DescribeTable("should pull the mount source stage missing in the local containers storage",
		func(ctx SpecContext, testOpts stagedDockerfileRunMountTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}
			buildahRuntime, ok := contRuntime.(*contback.NativeBuildahBackend)
			Expect(ok).To(BeTrue(), "test requires the native buildah backend")

			repoDirname := "repo0"

			By("state0: preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, "staged_dockerfile_run_mount/state0")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			By("state0: building images")
			Expect(werfProject.Build(ctx, nil)).To(ContainSubstring("Building stage"))

			By("state0: removing project images from the local containers storage")
			buildahRuntime.RmiByRepoRef(ctx, suite_init.TestRepo(SuiteData.ProjectName))

			By("state1: changing the final stage only")
			SuiteData.UpdateTestRepo(ctx, repoDirname, "staged_dockerfile_run_mount/state1")

			By("state1: rebuilding with the mount source stage present only in repo")
			buildOut := werfProject.Build(ctx, nil)
			Expect(buildOut).To(ContainSubstring("Use previously built image"))
			Expect(buildOut).To(ContainSubstring("Building stage"))
		},
		Entry("with local repo using Native Buildah with rootless isolation", stagedDockerfileRunMountTestOptions{setupEnvOptions{
			ContainerBackendMode: "native-rootless",
			WithLocalRepo:        true,
		}}),
		Entry("with local repo using Native Buildah with chroot isolation", stagedDockerfileRunMountTestOptions{setupEnvOptions{
			ContainerBackendMode: "native-chroot",
			WithLocalRepo:        true,
		}}),
	)
})
