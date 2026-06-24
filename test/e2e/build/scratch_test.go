package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/utils"
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
			_, err := contback.NewContainerBackend(testOpts.ContainerBackendMode)
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

			By("checking scratch image contents and labels")
			imageName := buildReport.Images["stapel-scratch"].DockerImageName
			utils.ExpectFileContentInImage(ctx, testOpts.ContainerBackendMode, imageName, "etc/werf-test-scratch-import", "werf-test-scratch-import\n")
			utils.ExpectImageHasNonEmptyLabels(ctx, testOpts.ContainerBackendMode, imageName,
				image.WerfLabel,
				image.WerfVersionLabel,
				image.WerfStageContentDigestLabel,
				image.WerfProjectRepoCommitLabel,
			)
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
