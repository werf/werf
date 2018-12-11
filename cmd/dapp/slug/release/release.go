package release

import (
	"fmt"

	"github.com/flant/dapp/pkg/slug"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release NAME",
		Args:  cobra.MinimumNArgs(1),
		Short: "Prints name suitable for Helm Release based on the specified NAME",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(slug.HelmRelease(args[0]))
		},
	}

	return cmd
}
