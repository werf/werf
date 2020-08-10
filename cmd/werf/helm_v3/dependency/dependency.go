package dependency

import (
	"os"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
)

func NewCmd() *cobra.Command {
	return helm_v3.NewDependencyCmd(os.Stdout)
}
