package namespace

import (
	"fmt"

	"github.com/flant/werf/pkg/slug"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace NAME",
		Args:  cobra.MinimumNArgs(1),
		Short: "Prints name suitable for Kubernetes Namespace based on the specified NAME",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(slug.KubernetesNamespace(args[0]))
		},
	}

	return cmd
}
