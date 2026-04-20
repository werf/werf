package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/werf"
)

const sbomEmulationWarning = "WARNING: SBOM generation is running in emulation mode, skipping actual generation"

var _ = Describe("Sbom get", Label("e2e", "sbom", "get", "simple"), func() {
	Describe("default", func() {
		DescribeTable("should succeed with SBOM emulation",
			func(ctx SpecContext, testOpts simpleTestOptions) {
				By("initializing")
				setupEnv(testOpts.setupEnvOptions)

				repoDirname := "repo0"
				fixtureRelPath := "state0"

				By("preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

				By("building images")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				buildOut := werfProject.Build(ctx, nil)
				Expect(buildOut).To(ContainSubstring(sbomEmulationWarning))
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
			XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
			XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "native-chroot",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
		)
	})

	Describe("lightweight", Label("tag"), func() {
		DescribeTable("should succeed with SBOM emulation (tag)",
			func(ctx SpecContext, testOpts simpleTestOptions) {
				By("initializing")
				setupEnv(testOpts.setupEnvOptions)

				repoDirname := "repo0-tag"

				By("preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, "state0")

				By("building images")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				buildOut := werfProject.Build(ctx, nil)
				Expect(buildOut).To(ContainSubstring(sbomEmulationWarning))
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
			XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
			XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "native-chroot",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
		)
	})

	Describe("lightweight", Label("digest"), func() {
		DescribeTable("should succeed with SBOM emulation (digest)",
			func(ctx SpecContext, testOpts simpleTestOptions) {
				By("initializing")
				setupEnv(testOpts.setupEnvOptions)

				repoDirname := "repo0-digest"

				By("preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, "state0")

				By("building images")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				buildOut := werfProject.Build(ctx, nil)
				Expect(buildOut).To(ContainSubstring(sbomEmulationWarning))
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
			XEntry("with local repo using Native Buildah with rootless isolation", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "native-rootless",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
			XEntry("with local repo using Native Buildah with chroot isolation", simpleTestOptions{setupEnvOptions{
				ContainerBackendMode:        "native-chroot",
				WithLocalRepo:               true,
				WithStagedDockerfileBuilder: false,
			}}),
		)
	})

	Describe("negative cases", Label("negative"), func() {
		DescribeTable("should fail with mutually exclusive flags",
			func(ctx SpecContext, extraArgs []string) {
				setupEnv(setupEnvOptions{
					ContainerBackendMode: "vanilla-docker",
					WithLocalRepo:        true,
				})

				repoDirname := "repo0-neg"
				SuiteData.InitTestRepo(ctx, repoDirname, "state0")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				werfProject.SbomGet(ctx, &werf.SbomGetOptions{
					CommonOptions: werf.CommonOptions{
						ShouldFail: true,
						ExtraArgs:  extraArgs,
					},
				})
			},
			Entry("--tag and --digest together",
				[]string{
					"--tag", "some-tag",
					"--digest", "sha256:abc123",
					"--repo", "localhost:5000/test",
				},
			),
		)
	})
})
