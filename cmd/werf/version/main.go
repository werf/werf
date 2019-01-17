package version

import (
	"fmt"

	"github.com/flant/werf/pkg/werf"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(werf.Version)
		},
	}

	return cmd
}
