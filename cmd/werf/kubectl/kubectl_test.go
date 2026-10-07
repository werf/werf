package kubectl

import (
	"slices"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "kubectl suite")
}

// fakePluginHandler serves the plugins listed in installed and records the
// executed plugin instead of exec'ing a binary.
type fakePluginHandler struct {
	installed []string
	executed  string
	execArgs  []string
}

func (h *fakePluginHandler) Lookup(filename string) (string, bool) {
	if slices.Contains(h.installed, filename) {
		return "/usr/local/bin/kubectl-" + filename, true
	}

	return "", false
}

func (h *fakePluginHandler) Execute(path string, args, _ []string) error {
	h.executed, h.execArgs = path, args
	return nil
}

var _ = Describe("kubectl plugin lookup", func() {
	DescribeTable("pluginLookupArgs",
		func(args []string, selfInvocationCommand string, expected []string) {
			Expect(pluginLookupArgs(args, selfInvocationCommand)).To(Equal(expected))
		},
		Entry("werf kubectl drops the command path",
			[]string{"werf", "kubectl", "argo", "rollouts", "get"}, "",
			[]string{"werf", "argo", "rollouts", "get"}),
		Entry("bare werf kubectl",
			[]string{"werf", "kubectl"}, "",
			[]string{"werf"}),
		Entry("other werf command disables the lookup",
			[]string{"werf", "converge", "--dev"}, "",
			[]string{"werf"}),
		Entry("bare werf",
			[]string{"werf"}, "",
			[]string{"werf"}),
		Entry("empty argv",
			[]string{}, "",
			[]string{}),
		Entry("embedded werf kubectl drops the self-invocation command",
			[]string{"d8", "dk", "kubectl", "argo", "rollouts"}, "dk",
			[]string{"d8", "argo", "rollouts"}),
		Entry("embedded werf, other command disables the lookup",
			[]string{"d8", "dk", "converge"}, "dk",
			[]string{"d8"}),
		Entry("embedding binary command disables the lookup",
			[]string{"d8", "k", "argo", "rollouts"}, "dk",
			[]string{"d8"}),
		Entry("multi-word self-invocation command",
			[]string{"host", "tools", "werf", "kubectl", "hello"}, "tools werf",
			[]string{"host", "hello"}),
	)

	DescribeTable("runs only plugins requested through kubectl",
		func(args []string, expectedExecuted string, expectedArgs []string) {
			h := &fakePluginHandler{installed: []string{"argo-rollouts"}}

			cmd.NewDefaultKubectlCommandWithArgs(cmd.KubectlOptions{
				PluginHandler: h,
				Arguments:     pluginLookupArgs(args, ""),
				IOStreams:     genericclioptions.NewTestIOStreamsDiscard(),
			})

			Expect(h.executed).To(Equal(expectedExecuted))
			Expect(h.execArgs).To(Equal(expectedArgs))
		},
		Entry("werf kubectl runs a multi-word plugin",
			[]string{"werf", "kubectl", "argo", "rollouts", "get", "rollout", "demo"},
			"/usr/local/bin/kubectl-argo-rollouts", []string{"get", "rollout", "demo"}),
		Entry("other werf command never runs a kubectl plugin",
			[]string{"werf", "argo", "rollouts"},
			"", nil),
	)
})
