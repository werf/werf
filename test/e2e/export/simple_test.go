package e2e_export_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Simple export", Label("e2e", "export", "simple"), func() {
	It("should succeed and export images",
		func() {
			SuiteData.WerfRepo = strings.Join([]string{SuiteData.RegistryLocalAddress, SuiteData.ProjectName}, "/")
			By("state0 running")
			{
				repoDirname := "repo0"
				fixtureRelPath := "simple/state0"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By("state0: running export")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				exportOut := werfProject.Export(&werf.ExportOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{
							"--dev",
							"--tag",
							SuiteData.WerfRepo,
						},
					},
				})
				Expect(exportOut).To(ContainSubstring("Exporting image..."))
			}
			By("state1 running")
			{
				repoDirname := "repo0"
				fixtureRelPath := "simple/state1"

				By("state1: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By("state2: running export multiplatform image")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				SuiteData.Stubs.SetEnv("WERF_REPO", SuiteData.WerfRepo)
				exportOut := werfProject.Export(&werf.ExportOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{
							"--dev",
							"--tag",
							SuiteData.WerfRepo,
						},
					},
				})
				Expect(exportOut).To(ContainSubstring("Exporting image..."))
			}
		})
})
