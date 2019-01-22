package namespace

import (
	"fmt"

	"github.com/flant/werf/pkg/slug"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "namespace NAME",
		DisableFlagsInUseLine: true,
		Args:  cobra.MinimumNArgs(1),
		Short: "Prints name suitable for Kubernetes Namespace based on the specified NAME",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(slug.KubernetesNamespace(args[0]))
		},
		Example: `  $ werf slug namespace feature-fix-2
  feature-fix-2

  $ werf slug namespace 'branch/one/!@#4.4-3'
  branch-one-4-4-3-4fe08955

  $ werf slug namespace My_branch
  my-branch-8ebf2d1d`,
	}

	return cmd
}
