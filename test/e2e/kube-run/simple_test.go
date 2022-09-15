package e2e_kube_run_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/test/pkg/werf"
)

var _ = Describe("Simple kube-run", Label("e2e", "kube-run", "simple"), func() {
	DescribeTable("should",
		func(kubeRunOpts *werf.KubeRunOptions, expectOutFn func(out string)) {
			By("initializing")
			setupEnv()
			repoDirname := "repo0"
			fixtureRelPath := "simple/state0"

			By("state0: preparing test repo")
			SuiteData.InitTestRepo(repoDirname, fixtureRelPath)
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			By("state0: execute kube-run")
			combinedOut := werfProject.KubeRun(kubeRunOpts)
			expectOutFn(combinedOut)
		},
		Entry(
			"succeed and produce expected output, running non-interactively",
			&werf.KubeRunOptions{
				Command: []string{"sh", "-ec", "cat /etc/os-release"},
				Image:   "main",
			},
			func(out string) {
				Expect(out).To(ContainSubstring("ID=ubuntu"))
			},
		),
		Entry(
			"succeed and produce expected output, running interactively with TTY",
			&werf.KubeRunOptions{
				Command: []string{"sh", "-ec", "cat /etc/os-release"},
				Image:   "main",
				CommonOptions: werf.CommonOptions{
					ExtraArgs: []string{"-i", "-t"},
				},
			},
			func(out string) {
				Expect(out).To(ContainSubstring("ID=ubuntu"))
			},
		),
		Entry(
			"fail and produce expected output, running non-interactively",
			&werf.KubeRunOptions{
				Command: []string{"sh", "-ec", "cat /etc/os-release; exit 1"},
				Image:   "main",
				CommonOptions: werf.CommonOptions{
					ShouldFail: true,
				},
			},
			func(out string) {
				Expect(out).To(ContainSubstring("ID=ubuntu"))
			},
		),
		Entry(
			"fail and produce expected output, running interactively with TTY",
			&werf.KubeRunOptions{
				Command: []string{"sh", "-ec", "cat /etc/os-release; exit 1"},
				Image:   "main",
				CommonOptions: werf.CommonOptions{
					ShouldFail: true,
					ExtraArgs:  []string{"-i", "-t"},
				},
			},
			func(out string) {
				Expect(out).To(ContainSubstring("ID=ubuntu"))
			},
		),
	)
})

func setupEnv() {
	SuiteData.Stubs.SetEnv("WERF_REPO", strings.Join([]string{os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY"), SuiteData.ProjectName}, "/"))

	if util.GetBoolEnvironmentDefaultFalse("WERF_TEST_K8S_DOCKER_REGISTRY_INSECURE") {
		SuiteData.Stubs.SetEnv("WERF_INSECURE_REGISTRY", "1")
		SuiteData.Stubs.SetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY", "1")
	} else {
		SuiteData.Stubs.UnsetEnv("WERF_INSECURE_REGISTRY")
		SuiteData.Stubs.UnsetEnv("WERF_SKIP_TLS_VERIFY_REGISTRY")
	}
}
