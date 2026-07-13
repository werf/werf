package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type simpleTestOptions struct {
	setupEnvOptions
}

var _ = Describe("Simple build", Label("e2e", "build", "simple"), func() {
	DescribeTable("should succeed and produce expected image",
		func(ctx SpecContext, testOpts simpleTestOptions) {
			By("initializing")
			setupEnv(testOpts.setupEnvOptions)
			contRuntime := contback.NewContainerBackend()

			By("state0: starting")
			{
				repoDirname := "repo0"
				fixtureRelPath := "simple/state0"
				buildReportName := "report0.json"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)

				By("state0: building images")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				reportProject := report.NewProjectWithReport(werfProject)
				buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), nil)
				Expect(buildOut).To(ContainSubstring("Building stage"))
				Expect(buildOut).NotTo(ContainSubstring("Use previously built image"))

				By("state0: rebuilding same images")
				Expect(werfProject.Build(ctx, nil)).To(And(
					ContainSubstring("Use previously built image"),
					Not(ContainSubstring("Building stage")),
				))

				By(`state0: checking "dockerfile" image content`)
				contRuntime.ExpectCmdsToSucceed(
					ctx,
					buildReport.Images["dockerfile"].DockerImageName,
					"test -f /file",
					"echo 'filecontent' | diff /file -",

					"test -f /created-by-run",
				)

				By(`state0: checking "stapel-shell" image content`)
				contRuntime.ExpectCmdsToSucceed(
					ctx,
					buildReport.Images["stapel-shell"].DockerImageName,
					"test -f /file",
					"stat -c %u:%g /file | diff <(echo 0:0) -",
					"echo 'filecontent' | diff /file -",

					"test -f /created-by-setup",
				)
			}
		},
		Entry("with local repo using BuildKit", simpleTestOptions{setupEnvOptions{
			WithStagedDockerfileBuilder: false,
		}}),
	)
})
