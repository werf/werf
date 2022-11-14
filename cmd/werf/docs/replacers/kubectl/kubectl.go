package kubectl

import (
	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/common"
)

func ReplaceKubectlDocs(cmd *cobra.Command) *cobra.Command {
	if len(cmd.Commands()) > 0 {
		for _, c := range cmd.Commands() {
			if len(c.Commands()) > 0 {
				ReplaceKubectlDocs(c)
			}
			setNewDocs(c)
		}
	}
	setNewDocs(cmd)
	return cmd
}

func setNewDocs(cmd *cobra.Command) {
	switch cmd.Use {
	case "events [--for TYPE/NAME] [--watch]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAlphaEventsDocs().LongMD,
		}
	case "kubectl":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetKubectlDocs().LongMD,
		}
	case "alpha":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAlphaDocs().LongMD,
		}
	case "annotate [--overwrite] (-f FILENAME | TYPE NAME) KEY_1=VAL_1 ... KEY_N=VAL_N [--resource-version=version]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAnnotateDocs().LongMD,
		}
	case "api-resources":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApiResourcesDocs().LongMD,
		}
	case "api-versions":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApiVersionsDocs().LongMD,
		}
	}
}
