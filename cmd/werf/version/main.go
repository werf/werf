package version

import (
	"fmt"

	"github.com/werf/werf/pkg/werf"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "version",
		DisableFlagsInUseLine: true,
		Short:                 "Print version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(werf.Version)
		},
	}

	return cmd
}
