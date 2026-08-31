package e2e_build_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/contback"
	"github.com/werf/werf/v2/test/pkg/report"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Staged dockerfile COPY flags", Label("e2e", "build", "copy-flags"), func() {
	It("keeps source paths under the destination with --parents and drops --exclude matches", func(ctx SpecContext) {
		By("initializing")
		setupEnv(setupEnvOptions{
			ContainerBackendMode:        "native-rootless",
			WithLocalRepo:               true,
			WithStagedDockerfileBuilder: true,
		})
		contRuntime, err := contback.NewContainerBackend("native-rootless")
		if errors.Is(err, contback.ErrRuntimeUnavailable) {
			Skip(err.Error())
		} else if err != nil {
			Fail(err.Error())
		}

		By("preparing test repo")
		SuiteData.InitTestRepo(ctx, "repo0", "copy_flags")

		By("building image")
		werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath("repo0"))
		reportProject := report.NewProjectWithReport(werfProject)
		buildOut, buildReport := reportProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath("report0.json"), nil)
		Expect(buildOut).To(ContainSubstring("Building stage"))

		By("checking image content")
		contRuntime.ExpectCmdsToSucceed(ctx, buildReport.Images["dockerfile"].DockerImageName,
			"test -f /target/aaa/x",
			"test -f /target/bbb/y",
			"test ! -e /target/x",
			"test -f /pivot/x",
			"test ! -e /pivot/aaa",
			"test -f /single/aaa/x",
			"test -f /docs/keep.txt",
			"test -f /docs/sub/keep.txt",
			"test -f /docs/sub/skip.md",
			"test ! -e /docs/skip.md",
			"test -f /docs-deep/keep.txt",
			"test ! -e /docs-deep/skip.md",
			"test ! -e /docs-deep/sub/skip.md",
		)
	})
})
