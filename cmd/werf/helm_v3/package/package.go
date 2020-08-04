package package_

import (
	"os"

	helm_v3 "helm.sh/helm/v3/cmd/helm"

	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	return helm_v3.NewPackageCmd(os.Stdout)
}
