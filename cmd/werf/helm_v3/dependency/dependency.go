package dependency

import (
	"os"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
)

// NOTE: dependency commands does not need action config nor namespace
func NewCmd() *cobra.Command {
	//actionConfig := new(action.Configuration)
	cmd := helm_v3.NewDependencyCmd(os.Stdout)
	//for _, subCmd := range cmd.Commands() {
	//	common.SetupCmdNamespace(subCmd, actionConfig)
	//}
	return cmd
}
