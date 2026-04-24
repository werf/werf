package list

import (
	"cmp"
	"context"
	"fmt"
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/nelm/pkg/action"
	"github.com/werf/werf/v2/cmd/werf/common"
)

var cmdData struct {
	OutputFormat string
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "list",
		Short:                 "List deployed releases",
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := commonCmdData.ProcessFlags(); err != nil {
				return err
			}

			ctx = action.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultReleaseListLogLevel), action.SetupLoggingOptions{ColorMode: lo.FromPtr(commonCmdData.LogColorMode)})

			if _, err := action.ReleaseList(ctx, action.ReleaseListOptions{
				KubeConnectionOptions:       commonCmdData.KubeConnectionOptions,
				NetworkParallelism:          commonCmdData.NetworkParallelism,
				OutputFormat:                cmdData.OutputFormat,
				ReleaseNamespace:            commonCmdData.Namespace,
				ReleaseStorageDriver:        commonCmdData.ReleaseStorageDriver,
				ReleaseStorageSQLConnection: commonCmdData.ReleaseStorageSQLConnection,
				TempDirPath:                 lo.FromPtr(commonCmdData.TmpDir),
			}); err != nil {
				return fmt.Errorf("release list: %w", err)
			}

			return nil
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	lo.Must0(common.SetupKubeConnectionFlags(&commonCmdData, cmd))
	common.SetupNamespace(&commonCmdData, cmd, false)
	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupReleaseStorageDriver(&commonCmdData, cmd)
	common.SetupReleaseStorageSQLConnection(&commonCmdData, cmd)
	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.OutputFormat, "output-format", "", cmp.Or(os.Getenv("WERF_OUTPUT_FORMAT"), action.DefaultReleaseListOutputFormat), "Output format. Options: table, yaml, json (default $WERF_OUTPUT_FORMAT or \""+action.DefaultReleaseListOutputFormat+"\")")

	return cmd
}
