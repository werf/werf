package tag

import (
	"fmt"

	"github.com/flant/werf/pkg/slug"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "tag NAME",
		DisableFlagsInUseLine: true,
		Args:  cobra.MinimumNArgs(1),
		Short: "Prints name suitable for Docker Tag based on the specified NAME",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(slug.DockerTag(args[0]))
		},
		Example: `  $ werf slug tag helo/ehlo
  helo-ehlo-b6f6ab1f

  $ werf slug tag 16.04
  16.04`,
	}

	return cmd
}
