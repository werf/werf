package helm_v3

import (
	"os"

	"github.com/spf13/cobra"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"

	"helm.sh/helm/v3/pkg/action"

	"github.com/werf/werf/pkg/deploy/helm_v3"
	cmd_helm "helm.sh/helm/v3/cmd/helm"
)

func NewInstallCmd(actionConfig *action.Configuration, commonCmdData *cmd_werf_common.CmdData) *cobra.Command {
	postRenderer := helm_v3.NewExtraAnnotationsAndLabelsPostRenderer(helm_v3.DefaultExtraAnnotations, nil)
	cmd := cmd_helm.NewInstallCmd(actionConfig, os.Stdout, cmd_helm.InstallCmdOptions{PostRenderer: postRenderer})
	SetupExtraAnnotationsAndLabelsForCmd(cmd, commonCmdData, postRenderer)
	return cmd
}
