package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/report"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = Describe("Scratch stapel build", Label("e2e", "build", "scratch", "simple"), func() {
	DescribeTable("should build scratch stapel image",
		func(ctx SpecContext, testOpts setupEnvOptions) {
			By("initializing")
			setupEnv(testOpts)
			contback.SkipIfUnavailable(testOpts.ContainerBackendMode)

			repoDirname := "repo0"
			fixtureRelPath := "scratch/state0"
			buildReportName := "report0.json"

			By("preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

			By("building image")
			werfProject := newWerfProject(repoDirname)
			reportProject := report.NewProjectWithReport(werfProject)
			buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)
			Expect(buildOut).To(ContainSubstring("Building stage"))

			By("checking scratch image contents and labels")
			imageName := buildReport.Images["stapel-scratch"].DockerImageName
			utils.ExpectImageIsReadable(ctx, testOpts.ContainerBackendMode, imageName)
			utils.ExpectFileContentInImage(ctx, testOpts.ContainerBackendMode, imageName, "etc/werf-test-scratch-import", "werf-test-scratch-import\n")
			utils.ExpectImageHasNonEmptyLabels(ctx, testOpts.ContainerBackendMode, imageName,
				image.WerfLabel,
				image.WerfVersionLabel,
				image.WerfStageContentDigestLabel,
				image.WerfProjectRepoCommitLabel,
			)
		},
		backendEntry("without repo using Docker", setupEnvOptions{
			ContainerBackendMode:        "docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}),
		backendEntry("with local repo using Docker", setupEnvOptions{
			ContainerBackendMode:        "docker",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
		backendEntry("with local repo using Native Buildah with rootless isolation", setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: false,
		}),
	)
})
