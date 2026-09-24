package history

import (
	"cmp"
	"context"
	"fmt"
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/nelm/v2/pkg/action"
	"github.com/werf/werf/v2/cmd/werf/common"
)

var cmdData struct {
	OutputFormat   string
	RevisionsLimit int
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
				RevisionsLimit:              cmdData.RevisionsLimit,
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
	revisionsLimit, err := util.GetIntEnvVar("WERF_RELEASE_HISTORY_REVISIONS_LIMIT")
	if err != nil {
		panic(fmt.Sprintf("bad WERF_RELEASE_HISTORY_REVISIONS_LIMIT value: %s", err))
	}
	cmd.Flags().IntVarP(&cmdData.RevisionsLimit, "revisions-limit", "", int(lo.FromPtr(revisionsLimit)), "Maximum number of revisions to show. 0 means no limit (default $WERF_RELEASE_HISTORY_REVISIONS_LIMIT or 0)")

	return cmd
}
