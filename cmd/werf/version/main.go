package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/pkg/werf"
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
