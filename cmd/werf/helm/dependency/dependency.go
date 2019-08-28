package dependency

import (
	"github.com/spf13/cobra"
)

func NewDependencyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "dependency update|build|list",
		Short:                 "Manage a chart's dependencies",
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(newDependencyListCmd())
	cmd.AddCommand(newDependencyUpdateCmd())
	cmd.AddCommand(newDependencyBuildCmd())

	return cmd
}
