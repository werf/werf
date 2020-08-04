package template

import (
	"os"

	"helm.sh/helm/v3/pkg/action"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
)

func NewCmd() *cobra.Command {
	actionConfig := new(action.Configuration)
	cmd := helm_v3.NewTemplateCmd(actionConfig, os.Stdout)
	return cmd
}
