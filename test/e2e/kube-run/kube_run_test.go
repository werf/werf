package e2e_kube_run_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/werf"
)

var _ = Describe("kube-run", func() {
	DescribeTable("should succeed/fail and produce expected output",
		func(kubeRunOpts *werf.KubeRunOptions, outputExpectationsFunc func(out string)) {
			repoDirname := "repo0"
			fixtureRelPath := "state0"

			By("preparing test repo")
			SuiteData.InitTestRepo(repoDirname, fixtureRelPath)
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			By("execute kube-run")
			combinedOut := werfProject.KubeRun(kubeRunOpts)
			outputExpectationsFunc(combinedOut)
		},
		Entry(
			"show output and succeed, running non-interactively",
			&werf.KubeRunOptions{
				Command: `sh -c "cat /etc/os-release"`,
				Image:   "main",
			},
			func(out string) {
				Expect(out).To(ContainSubstring("ID=alpine"))
			},
		),
		Entry(
			"show output and succeed, running interactively with TTY",
			&werf.KubeRunOptions{
				Command: `sh -c "cat /etc/os-release"`,
				Image:   "main",
				CommonOptions: werf.CommonOptions{
					ExtraArgs: []string{"-i", "-t"},
				},
			},
			func(out string) {
				Expect(out).To(ContainSubstring("ID=alpine"))
			},
		),
		Entry(
			"show output and fail, running non-interactively",
			&werf.KubeRunOptions{
				Command: `sh -c "cat /etc/os-release; exit 1"`,
				Image:   "main",
				CommonOptions: werf.CommonOptions{
					ShouldFail: true,
				},
			},
			func(out string) {
				Expect(out).To(ContainSubstring("ID=alpine"))
			},
		),
		Entry(
			"show output and fail, running interactively with TTY",
			&werf.KubeRunOptions{
				Command: `sh -c "cat /etc/os-release; exit 1"`,
				Image:   "main",
				CommonOptions: werf.CommonOptions{
					ShouldFail: true,
					ExtraArgs:  []string{"-i", "-t"},
				},
			},
			func(out string) {
				Expect(out).To(ContainSubstring("ID=alpine"))
			},
		),
	)
})
