package lint

import (
	"os"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
)

func NewCmd() *cobra.Command {
	cmd := helm_v3.NewLintCmd(os.Stdout)
	return cmd
}
