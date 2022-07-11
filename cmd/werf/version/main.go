package version

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/werf"
)

func NewCmd(ctx context.Context) *cobra.Command {
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "version",
		DisableFlagsInUseLine: true,
		Short:                 "Print version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(werf.Version)
		},
	})

	return cmd
}
