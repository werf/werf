package e2e_compose_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Complex compose", Label("e2e", "compose", "complex"), func() {
	It("should succeed and deploy expected resources",
		func() {
			By("state0: starting")
			{
				repoDirname := "repo0"
				fixtureRelPath := "simple/state0"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By("state0: running compose up with no options")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				composeOut := werfProject.Compose(&werf.BuildOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{"up"},
					},
				})
				Expect(composeOut).To(ContainSubstring("Building stage"))
				By("state0: running compose down")
				composeDownOut := werfProject.Compose(&werf.BuildOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{"down"},
					},
				})
				Expect(composeDownOut).To(ContainSubstring("Removed"))
			}
			By("state1: starting")
			{
				repoDirname := "repo0"
				fixtureRelPath := "simple/state1"
				By("state1: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By("state1: running compose with multiple compose files and args")
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				composeOut := werfProject.Compose(&werf.BuildOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{
							"up",
							"--docker-compose-options", "-f docker-compose.yaml -f docker-compose-b.yaml",
							"--docker-compose-command-options", "--always-recreate-deps"},
					},
				})
				Expect(composeOut).To(ContainSubstring("image backend"))
			}
		},
	)
})
