package e2e_compose_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/werf"
)

const (
	commandUp   = "up"
	commandDown = "down"
)

type simpleTestOptions struct {
	ExtraArgs        []string
	State            string
	StateDescription string
	Repo             string
}

var _ = Describe("Complex compose", Label("e2e", "compose", "complex"), func() {
	DescribeTable("should",
		func(opts simpleTestOptions) {
			By(fmt.Sprintf("%s: starting", opts.State))
			{
				repoDirname := opts.Repo
				fixtureRelPath := fmt.Sprintf("simple/%s", opts.State)

				By(fmt.Sprintf("%s: preparing test repo", opts.State))
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By(fmt.Sprintf("%s: %s", opts.State, opts.StateDescription))
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				composeOut := werfProject.Compose(&werf.BuildOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: append([]string{commandUp}, opts.ExtraArgs...),
					},
				})
				Expect(composeOut).To(ContainSubstring("Building stage"))
				Expect(composeOut).To(ContainSubstring("image backend"))
				By(fmt.Sprintf("%s: running compose down", opts.State))
				composeDownOut := werfProject.Compose(&werf.BuildOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: []string{commandDown},
					},
				})
				Expect(composeDownOut).To(ContainSubstring("Removed"))
			}
		},
		Entry("without additional options", simpleTestOptions{
			ExtraArgs:        []string{},
			State:            "state0",
			StateDescription: "running compose up with no options",
			Repo:             "repo0",
		}),
		Entry("with multiple compose files", simpleTestOptions{
			ExtraArgs: []string{
				"--docker-compose-options", "-f docker-compose.yaml -f docker-compose-b.yaml",
				"--docker-compose-command-options", "--always-recreate-deps",
			},
			State:            "state1",
			StateDescription: "running compose up with multiple compose files and args",
			Repo:             "repo1",
		}),
	)
})
