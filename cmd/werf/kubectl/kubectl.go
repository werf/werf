package kubectl

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/cmd"
)

func NewCmd() *cobra.Command {
	kubectlRootCmd := cmd.NewDefaultKubectlCommand()
	return kubectlRootCmd
}
