package release

import (
	"fmt"

	"github.com/flant/werf/pkg/slug"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "release NAME",
		DisableFlagsInUseLine: true,
		Args:  cobra.MinimumNArgs(1),
		Short: "Prints name suitable for Helm Release based on the specified NAME",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(slug.HelmRelease(args[0]))
		},
		Example: `  $ werf slug release my_release-NAME
  my_release-NAME

  The result has been trimmed to fit maximum bytes limit:
  $ werf slug release looooooooooooooooooooooooooooooooooooooooooong_string
  looooooooooooooooooooooooooooooooooooooooooong-stri-b150a895`,
	}

	return cmd
}
