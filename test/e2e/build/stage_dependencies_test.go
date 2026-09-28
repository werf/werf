package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/test/pkg/contback"
	"github.com/werf/werf/v3/test/pkg/report"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = Describe("Default stage dependencies", Label("e2e", "build", "stage-dependencies"), func() {
	DescribeTable("should preserve renamed files instead of reusing a stale content anchor",
		func(ctx SpecContext, testOpts setupEnvOptions) {
			setupEnv(testOpts)
			contRuntime := contback.NewContainerBackend(testOpts.ContainerBackendMode)
			repoDirname := "repo0"
			SuiteData.InitTestRepo(ctx, repoDirname, "stage_dependencies/state0")
			repoPath := SuiteData.GetTestRepoPath(repoDirname)
			werfProject := newWerfProject(repoDirname)
			reportProject := report.NewProjectWithReport(werfProject)

			By("building the original file")
			_, before := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("before.json"), nil)
			contRuntime.ExpectCmdsToSucceed(ctx, before.Images["app"].DockerImageName,
				"test -f /app/app.sh", "test ! -e /app/renamed.sh")

			By("renaming without changing content or traversal order")
			utils.RunSucceedCommand(ctx, repoPath, "git", "mv", "app.sh", "renamed.sh")
			utils.RunSucceedCommand(ctx, repoPath, "git", "commit", "-m", "rename")

			By("building and checking the resulting filesystem")
			_, after := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("after.json"), nil)
			contRuntime.ExpectCmdsToSucceed(ctx, after.Images["app"].DockerImageName,
				"test ! -e /app/app.sh", "test -f /app/renamed.sh",
				"test \"$(bash /app/renamed.sh)\" = hello")
		},
		backendEntry("with local repo using Docker", setupEnvOptions{
			ContainerBackendMode: "docker",
			WithLocalRepo:        true,
		}),
	)

	DescribeTable("should rebuild stages with shell commands when source files change",
		func(ctx SpecContext, testOpts setupEnvOptions) {
			By("initializing")
			setupEnv(testOpts)
			contback.SkipIfUnavailable(testOpts.ContainerBackendMode)

			repoDirname := "repo0"

			By("state0: preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, "stage_dependencies/state0")

			By("state0: building images")
			werfProject := newWerfProject(repoDirname)
			buildOut := werfProject.Build(ctx, nil)
			Expect(buildOut).To(ContainSubstring("Building stage"))
			Expect(buildOut).NotTo(ContainSubstring("Use previously built image"))

			By("state0: rebuilding without changes reuses cache")
			Expect(werfProject.Build(ctx, nil)).To(And(
				ContainSubstring("Use previously built image"),
				Not(ContainSubstring("Building stage")),
			))

			By("state1: updating repo with changed source file")
			SuiteData.UpdateTestRepo(ctx, repoDirname, "stage_dependencies/state1")

			By("state1: rebuilding triggers stage rebuild due to default stageDependencies")
			Expect(werfProject.Build(ctx, nil)).To(ContainSubstring("Building stage"))
		},
		backendEntry("without repo using Docker", setupEnvOptions{
			ContainerBackendMode:        "docker",
			WithLocalRepo:               false,
			WithStagedDockerfileBuilder: false,
		}),
	)
})
