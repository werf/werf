package version

import (
	"fmt"

	"github.com/flant/dapp/pkg/dapp"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(dapp.Version)
		},
	}

	return cmd
}
