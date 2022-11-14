package kubectl

import (
	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/common"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/plugin"
	"os"
	"testing"
)

var (
	configFlags *genericclioptions.ConfigFlags

	result, textsResult      bool
	msgAnnotations, msgTexts string
)

func TestReplaceKubectlDocs(t *testing.T) {
	configFlags = genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	cmd := ReplaceKubectlDocs(cmd.NewDefaultKubectlCommandWithArgs(cmd.KubectlOptions{
		PluginHandler: cmd.NewDefaultPluginHandler(plugin.ValidPluginFilenamePrefixes),
		Arguments:     os.Args,
		ConfigFlags:   configFlags,
		IOStreams:     genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	}))

	msgAnnotations = "There are no annotations in the following commands:\n\n"
	msgTexts = "The following commands do not match the text in the annotation:\n\n"
	result = true
	textsResult = true
	checkCmd(cmd, "")
	if !result || !textsResult {
		if !result {
			t.Error(msgAnnotations + "\n\n")
		}
		if !textsResult {
			t.Error(msgTexts + "\n\n")
		}
	}
}

func checkCmd(cmd *cobra.Command, previuosCmd string) {
	if ann, ok := cmd.Annotations[common.DocsLongMD]; !ok {
		result = false
		msgAnnotations += previuosCmd + " " + cmd.Use + "\n\n"
	} else {
		if !checkText(ann) {
			textsResult = false
			msgTexts += previuosCmd + " " + cmd.Use + "\n\n"
		}
	}
	if len(cmd.Commands()) > 0 {
		for _, c := range cmd.Commands() {
			checkCmd(c, previuosCmd+" "+cmd.Use)
		}
	}
}

func checkText(ann string) bool {
	if ann != GetAlphaEventsDocs().LongMD &&
		ann != GetKubectlDocs().LongMD &&
		ann != GetAlphaDocs().LongMD &&
		ann != GetAnnotateDocs().LongMD &&
		ann != GetApiResourcesDocs().LongMD &&
		ann != GetApiVersionsDocs().LongMD &&
		ann != GetApplyDocs().LongMD &&
		ann != GetApplyEditLastAppliedDocs().LongMD &&
		ann != GetApplySetLastAppliedDocs().LongMD &&
		ann != GetApplyViewLastAppliedDocs().LongMD &&
		ann != GetAttachDocs().LongMD {
		return false
	}
	return true
}
