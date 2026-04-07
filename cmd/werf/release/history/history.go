package history

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
	Max          int
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "history",
		Short:                 "Show release history",
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

			releaseName, err := common.GetRequiredRelease(&commonCmdData)
			if err != nil {
				return err
			}

			releaseNamespace := common.GetNamespace(&commonCmdData)

			ctx = action.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultReleaseHistoryLogLevel), action.SetupLoggingOptions{ColorMode: lo.FromPtr(commonCmdData.LogColorMode)})

			if _, err := action.ReleaseHistory(ctx, releaseName, releaseNamespace, action.ReleaseHistoryOptions{
				KubeConnectionOptions:       commonCmdData.KubeConnectionOptions,
				Max:                         cmdData.Max,
				OutputFormat:                cmdData.OutputFormat,
				ReleaseStorageDriver:        commonCmdData.ReleaseStorageDriver,
				ReleaseStorageSQLConnection: commonCmdData.ReleaseStorageSQLConnection,
				TempDirPath:                 lo.FromPtr(commonCmdData.TmpDir),
			}); err != nil {
				return fmt.Errorf("release history: %w", err)
			}

			return nil
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	lo.Must0(common.SetupKubeConnectionFlags(&commonCmdData, cmd))
	common.SetupRelease(&commonCmdData, cmd, false)
	common.SetupNamespace(&commonCmdData, cmd, false)
	common.SetupReleaseStorageDriver(&commonCmdData, cmd)
	common.SetupReleaseStorageSQLConnection(&commonCmdData, cmd)
	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.OutputFormat, "output-format", "", cmp.Or(os.Getenv("WERF_OUTPUT_FORMAT"), action.DefaultReleaseHistoryOutputFormat), "Output format. Options: table, yaml, json (default $WERF_OUTPUT_FORMAT or \""+action.DefaultReleaseHistoryOutputFormat+"\")")
	cmd.Flags().IntVarP(&cmdData.Max, "max", "", 0, "Maximum number of revisions to show. 0 means no limit (default $WERF_RELEASE_HISTORY_MAX or 0)")

	return cmd
}
