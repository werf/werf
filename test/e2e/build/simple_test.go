package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/contback"
	"github.com/werf/werf/test/pkg/werf"
)

type simpleTestOptions struct {
	setupEnvOptions
}

var _ = Describe("Simple build", Label("e2e", "build", "simple"), func() {
	DescribeTable("should succeed and produce expected image",
		func(testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			By("state0: starting")
			{
				repoDirname := "repo0"
				fixtureRelPath := "simple/state0"
				buildReportName := "report0.json"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By("state0: building images")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				buildOut, buildReport := werfProject.BuildWithReport(SuiteData.GetBuildReportPath(buildReportName), nil)
				Expect(buildOut).To(ContainSubstring("Building stage"))
				Expect(buildOut).NotTo(ContainSubstring("Use previously built image"))

				By("state0: rebuilding same images")
				Expect(werfProject.Build(nil)).To(And(
					ContainSubstring("Use previously built image"),
					Not(ContainSubstring("Building stage")),
				))

				By(`state0: checking "dockerfile" image content`)
				contRuntime.ExpectCmdsToSucceed(
					buildReport.Images["dockerfile"].DockerImageName,
					"test -f /file",
					"echo 'filecontent' | diff /file -",

					"test -f /created-by-run",
				)

				By(`state0: checking "stapel-shell" image content`)
				contRuntime.ExpectCmdsToSucceed(
					buildReport.Images["stapel-shell"].DockerImageName,
					"test -f /file",
					"stat -c %u:%g /file | diff <(echo 0:0) -",
					"echo 'filecontent' | diff /file -",

					"test -f /created-by-setup",
				)
			}
		},
		Entry("without repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using Vanilla Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("without repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using BuildKit Docker", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "buildkit-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		// TODO(ilya-lesikov): uncomment after Staged Dockerfile builder finished
		// // TODO(1.3): after Full Dockerfile Builder removed and Staged Dockerfile Builder enabled by default this test no longer needed
		// Entry("with local repo using Native Buildah and Staged Dockerfile Builder with rootless isolation", simpleTestOptions{setupEnvOptions{
		// 	ContainerBackendMode:                 "native-rootless",
		// 	WithLocalRepo:               true,
		// 	WithStagedDockerfileBuilder: true,
		// }),
		// TODO(ilya-lesikov): uncomment after Staged Dockerfile builder finished
		// // TODO(1.3): after Full Dockerfile Builder removed and Staged Dockerfile Builder enabled by default this test no longer needed
		// Entry("with local repo using Native Buildah and Staged Dockerfile Builder with chroot isolation", simpleTestOptions{setupEnvOptions{
		// 	ContainerBackendMode:                 "native-chroot",
		// 	WithLocalRepo:               true,
		// 	WithStagedDockerfileBuilder: true,
		// }),
	)
})
