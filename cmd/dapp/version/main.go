package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/dapp/pkg/dapp"
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
