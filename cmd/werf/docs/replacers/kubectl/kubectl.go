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
	case "apply (-f FILENAME | -k DIRECTORY)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApplyDocs().LongMD,
		}
	case "edit-last-applied (RESOURCE/NAME | -f FILENAME)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApplyEditLastAppliedDocs().LongMD,
		}
	case "set-last-applied -f FILENAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApplySetLastAppliedDocs().LongMD,
		}
	case "view-last-applied (TYPE [NAME | -l label] | TYPE/NAME | -f FILENAME)":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetApplyViewLastAppliedDocs().LongMD,
		}
	case "attach (POD | TYPE/NAME) -c CONTAINER":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAttachDocs().LongMD,
		}
	case "auth":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAuthDocs().LongMD,
		}
	case "can-i VERB [TYPE | TYPE/NAME | NONRESOURCEURL]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAuthCanIDocs().LongMD,
		}
	case "reconcile -f FILENAME":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAuthReconcileDocs().LongMD,
		}
	case "autoscale (-f FILENAME | TYPE NAME | TYPE/NAME) [--min=MINPODS] --max=MAXPODS [--cpu-percent=CPU]":
		cmd.Annotations = map[string]string{
			common.DocsLongMD: GetAutoscaleDocs().LongMD,
		}
	}
}
