package e2e_kube_run_test

import (
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/test/pkg/werf"
)

var _ = Describe("Simple kube-run", Label("e2e", "kube-run", "simple"), func() {
	DescribeTable("should",
		func(ctx SpecContext, kubeRunOpts *werf.KubeRunOptions, expectOutFn func(out string)) {
			By("initializing")
			setupEnv()
			repoDirname := "repo0"
			fixtureRelPath := "simple/state0"

			By("state0: preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			By("state0: execute kube-run")
			combinedOut := werfProject.KubeRun(ctx, kubeRunOpts)
			Expect(combinedOut).To(ContainSubstring("Creating namespace"))
			Expect(combinedOut).To(ContainSubstring("Running pod"))
			Expect(combinedOut).To(ContainSubstring("Executing into pod"))
			Expect(combinedOut).To(ContainSubstring("Stopping container"))
			Expect(combinedOut).To(ContainSubstring("Cleaning up pod"))

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
		Entry(
			"should cancel long running process",
			&werf.KubeRunOptions{
				Command: []string{"/opt/app.sh"},
				Image:   "main",
				CommonOptions: werf.CommonOptions{
					ShouldFail:            true,
					ExtraArgs:             []string{}, // be able to work without "-it" options
					CancelOnOutput:        "Looping ...",
					CancelOnOutputTimeout: time.Minute,
				},
			},
			func(out string) {
				Expect(out).To(ContainSubstring("Signal container"))
				Expect(out).To(ContainSubstring("Signal handled"))   // from script
				Expect(out).To(ContainSubstring("Script completed")) // from script
			},
			SpecTimeout(time.Minute*3),
		),
	)

	DescribeTable("should",
		func(ctx SpecContext, kubeRunOpts *werf.KubeRunOptions, expectOutFn func(out string)) {
			By("initializing")
			setupEnv()
			repoDirname := "repo0"
			fixtureRelPath := "simple/state0"
			buildReportName := "report0.json"
			kubeRunOpts.ExtraArgs = append(kubeRunOpts.ExtraArgs, "--use-build-report", "--build-report-path", SuiteData.GetBuildReportPath(buildReportName))

			By("state0: preparing test repo")
			SuiteData.InitTestRepo(ctx, repoDirname, fixtureRelPath)
			werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

			By("state0: building images")
			buildOut, _ := werfProject.BuildWithReport(ctx, SuiteData.GetBuildReportPath(buildReportName), &werf.BuildWithReportOptions{
				CommonOptions: werf.CommonOptions{
					ShouldFail: false,
					ExtraArgs:  []string{"--save-build-report", "--build-report-path", SuiteData.GetBuildReportPath(buildReportName)},
				},
			})
			Expect(buildOut).To(ContainSubstring("Building stage"))
			Expect(buildOut).NotTo(ContainSubstring("Use previously built image"))

			By("state0: execute kube-run")
			combinedOut := werfProject.KubeRun(ctx, kubeRunOpts)
			Expect(combinedOut).To(ContainSubstring("Creating namespace"))
			Expect(combinedOut).To(ContainSubstring("Running pod"))
			Expect(combinedOut).To(ContainSubstring("Executing into pod"))
			Expect(combinedOut).To(ContainSubstring("Stopping container"))
			Expect(combinedOut).To(ContainSubstring("Cleaning up pod"))

			expectOutFn(combinedOut)
		},
		Entry(
			"succeed and produce expected output with using build report, running non-interactively",
			&werf.KubeRunOptions{
				Command: []string{"sh", "-ec", "cat /etc/os-release"},
				Image:   "main",
			},
			func(out string) {
				Expect(out).To(ContainSubstring("ID=ubuntu"))
			},
		),
		Entry(
			"succeed and produce expected output with using build report, running interactively with TTY",
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
			"fail and produce expected output with using build report, running non-interactively",
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
			"fail and produce expected output with using build report, running interactively with TTY",
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
		Entry(
			"should cancel long running process with using build report",
			&werf.KubeRunOptions{
				Command: []string{"/opt/app.sh"},
				Image:   "main",
				CommonOptions: werf.CommonOptions{
					ShouldFail:            true,
					ExtraArgs:             []string{}, // be able to work without "-it" options
					CancelOnOutput:        "Looping ...",
					CancelOnOutputTimeout: time.Minute,
				},
			},
			func(out string) {
				Expect(out).To(ContainSubstring("Signal container"))
				Expect(out).To(ContainSubstring("Signal handled"))   // from script
				Expect(out).To(ContainSubstring("Script completed")) // from script
			},
			SpecTimeout(time.Minute*3),
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
