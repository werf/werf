package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
)

var _ = Describe("Default stage dependencies", Label("e2e", "build", "stage-dependencies"), func() {
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
