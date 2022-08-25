package e2e_build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/contback"
	"github.com/werf/werf/test/pkg/werf"
)

var _ = Describe("Simple build", Label("e2e", "build", "simple"), func() {
	DescribeTable("should succeed and produce expected image",
		func(withLocalRepo bool, buildahMode string) {
			By("initializing")
			setupEnv(withLocalRepo, buildahMode)
			contRuntime, err := contback.NewContainerBackend(buildahMode)
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
				Expect(buildOut).NotTo(ContainSubstring("Use cache image"))

				By("state0: rebuilding same images")
				Expect(werfProject.Build(nil)).To(And(
					ContainSubstring("Use cache image"),
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
		Entry("without repo using Docker", false, "docker"),
		Entry("with local repo using Docker", true, "docker"),
		Entry("with local repo using Native Buildah with rootless isolation", true, "native-rootless"),
		Entry("with local repo using Native Buildah with chroot isolation", true, "native-chroot"),
		// TODO: uncomment when buildah allows building without --repo flag
		// Entry("with local repo using Native Buildah with rootless isolation", false, "native-rootless"),
		// Entry("with local repo using Native Buildah with chroot isolation", false, "native-chroot"),
	)
})
