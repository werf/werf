package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type scratchTestOptions struct {
	setupEnvOptions
}

var _ = Describe("Scratch stapel build", Label("e2e", "build", "scratch", "simple"), func() {
	DescribeTable("should build scratch stapel image",
		func(ctx SpecContext, testOpts scratchTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contRuntime, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			if err == contback.ErrRuntimeUnavailable {
				Skip(err.Error())
			} else if err != nil {
				Fail(err.Error())
			}

			repoDirname := "repo0"
			fixtureRelPath := "scratch/state0"
			buildReportName := "report0.json"

			By("preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building image")
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)
			Expect(buildOut).To(ContainSubstring("Building stage"))

			By("checking scratch image labels")
			imageName := buildReport.Images["stapel-scratch"].DockerImageName
			if testOpts.WithLocalRepo {
				contRuntime.Pull(ctx, imageName)
			}
			inspect := contRuntime.GetImageInspect(ctx, imageName)
			labels := inspect.Config.Labels
			Expect(labels).NotTo(BeNil())
			Expect(labels).To(HaveKey(image.WerfLabel))
			Expect(labels[image.WerfLabel]).NotTo(BeEmpty())
			Expect(labels).To(HaveKey(image.WerfVersionLabel))
			Expect(labels[image.WerfVersionLabel]).NotTo(BeEmpty())
			Expect(labels).To(HaveKey(image.WerfStageContentDigestLabel))
			Expect(labels[image.WerfStageContentDigestLabel]).NotTo(BeEmpty())
			Expect(labels).To(HaveKey(image.WerfProjectRepoCommitLabel))
			Expect(labels[image.WerfProjectRepoCommitLabel]).NotTo(BeEmpty())
		},
		Entry("using Vanilla Docker", scratchTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("using Vanilla Docker with local repo", scratchTestOptions{setupEnvOptions{
			ContainerBackendMode:        "vanilla-docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("using Native Buildah with chroot isolation", scratchTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}}),
		Entry("using Native Buildah with chroot isolation and local repo", scratchTestOptions{setupEnvOptions{
			ContainerBackendMode:        "native-chroot",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}}),
	)
})
