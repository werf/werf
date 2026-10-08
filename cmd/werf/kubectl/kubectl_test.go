package kubectl

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "kubectl suite")
}

var _ = ginkgo.DescribeTable("kubectl environment flags without installed plugins",
	func(kubeContext, skipTLS string, args, expected []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, os.Args[0], append([]string{"version"}, args...)...)
		command.Env = append(os.Environ(), "WERF_TEST_KUBECTL_CONFIG=1",
			"WERF_SELF_INVOCATION_COMMAND=", "WERF_KUBE_CONTEXT="+kubeContext,
			"WERF_SKIP_TLS_VERIFY_REGISTRY="+skipTLS, "PATH="+ginkgo.GinkgoT().TempDir())
		output, err := command.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "%s", output)
		var actual []string
		gomega.Expect(json.Unmarshal(output, &actual)).To(gomega.Succeed())
		gomega.Expect(actual).To(gomega.Equal(append(expected, "kubectl [flags] [options]")))
	},
	ginkgo.Entry("defaults", "", "", []string{}, []string{"", "false"}),
	ginkgo.Entry("context", "environment-context", "", []string{}, []string{"environment-context", "false"}),
	ginkgo.Entry("TLS", "", "true", []string{}, []string{"", "true"}),
	ginkgo.Entry("context and TLS", "environment-context", "true", []string{}, []string{"environment-context", "true"}),
	ginkgo.Entry("explicit flags override environment", "environment-context", "true",
		[]string{"--context=explicit-context", "--insecure-skip-tls-verify=false"}, []string{"explicit-context", "false"}),
)

var _ = ginkgo.Describe("kubectl plugin lookup", func() {
	ginkgo.DescribeTable("pluginLookupArgs",
		func(args []string, selfInvocationCommand string, expected []string) {
			gomega.Expect(pluginLookupArgs(args, selfInvocationCommand)).To(gomega.Equal(expected))
		},
		ginkgo.Entry("werf kubectl drops the command path",
			[]string{"werf", "kubectl", "argo", "rollouts", "get"}, "",
			[]string{"werf", "argo", "rollouts", "get"}),
		ginkgo.Entry("bare werf kubectl",
			[]string{"werf", "kubectl"}, "",
			[]string{"werf"}),
		ginkgo.Entry("other werf command disables the lookup",
			[]string{"werf", "converge", "--dev"}, "",
			[]string{"werf"}),
		ginkgo.Entry("bare werf",
			[]string{"werf"}, "",
			[]string{"werf"}),
		ginkgo.Entry("empty argv",
			[]string{}, "",
			[]string{}),
		ginkgo.Entry("embedded werf kubectl drops the self-invocation command",
			[]string{"d8", "dk", "kubectl", "argo", "rollouts"}, "dk",
			[]string{"d8", "argo", "rollouts"}),
		ginkgo.Entry("embedded werf, other command disables the lookup",
			[]string{"d8", "dk", "converge"}, "dk",
			[]string{"d8"}),
		ginkgo.Entry("embedding binary command disables the lookup",
			[]string{"d8", "k", "argo", "rollouts"}, "dk",
			[]string{"d8"}),
		ginkgo.Entry("multi-word self-invocation command",
			[]string{"host", "tools", "werf", "kubectl", "hello"}, "tools werf",
			[]string{"host", "hello"}),
	)

	ginkgo.DescribeTable("runs only plugins requested through kubectl",
		func(args []string, expectedExecuted string, expectedArgs []string) {
			h := &fakePluginHandler{installed: []string{"argo-rollouts"}}

			cmd.NewDefaultKubectlCommandWithArgs(cmd.KubectlOptions{
				PluginHandler: h,
				Arguments:     pluginLookupArgs(args, ""),
				IOStreams:     genericclioptions.NewTestIOStreamsDiscard(),
			})

			gomega.Expect(h.executed).To(gomega.Equal(expectedExecuted))
			gomega.Expect(h.execArgs).To(gomega.Equal(expectedArgs))
		},
		ginkgo.Entry("werf kubectl runs a multi-word plugin",
			[]string{"werf", "kubectl", "argo", "rollouts", "get", "rollout", "demo"},
			"/usr/local/bin/kubectl-argo-rollouts", []string{"get", "rollout", "demo"}),
		ginkgo.Entry("other werf command never runs a kubectl plugin",
			[]string{"werf", "argo", "rollouts"},
			"", nil),
	)
})

var _ = ginkgo.DescribeTable("NewCmd plugin execution",
	func(args []string, selfInvocationCommand, pluginName string, expectedArgs []string) {
		dir := ginkgo.GinkgoT().TempDir()
		if runtime.GOOS == "windows" {
			pluginName += ".exe"
		}
		copyPlugin(filepath.Join(dir, "kubectl-"+pluginName))
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, os.Args[0], args...)
		command.Env = append(os.Environ(), "WERF_TEST_KUBECTL_COMMAND=1",
			"WERF_SELF_INVOCATION_COMMAND="+selfInvocationCommand, "KUBECTL_ENABLE_CMD_SHADOW=",
			"PATH="+dir)
		output, err := command.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "%s", output)
		if expectedArgs == nil {
			gomega.Expect(string(output)).To(gomega.BeEmpty())
			return
		}
		var actualArgs []string
		gomega.Expect(json.Unmarshal(output, &actualArgs)).To(gomega.Succeed())
		gomega.Expect(actualArgs).To(gomega.Equal(expectedArgs))
	},
	ginkgo.Entry("standalone multi-word plugin", []string{"kubectl", "argo", "rollouts", "get", "demo"}, "", "argo-rollouts", []string{"get", "demo"}),
	ginkgo.Entry("create subcommand plugin", []string{"kubectl", "create", "hello"}, "", "create-hello", []string{}),
	ginkgo.Entry("create plugin lookup requires an exact name", []string{"kubectl", "create", "hello", "demo"}, "", "create-hello", []string(nil)),
	ginkgo.Entry("embedded plugin", []string{"dk", "kubectl", "argo", "rollouts", "demo"}, "dk", "argo-rollouts", []string{"demo"}),
	ginkgo.Entry("unrelated command", []string{"argo", "rollouts", "demo"}, "", "argo-rollouts", []string(nil)),
	ginkgo.Entry("embedding binary command", []string{"argo", "rollouts", "demo"}, "dk", "argo-rollouts", []string(nil)),
	ginkgo.Entry("builtin takes precedence", []string{"kubectl", "version"}, "", "version", []string(nil)),
)
