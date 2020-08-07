package get

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/helm_v3/common"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
)

func NewCmd() *cobra.Command {
	actionConfig := new(action.Configuration)
	cmd := helm_v3.NewGetCmd(actionConfig, os.Stdout)
	for _, subCmd := range cmd.Commands() {
		common.SetupCmdActionConfig(subCmd, actionConfig)
	}
	return cmd
}
