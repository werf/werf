package get

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/werf/v2/cmd/werf/common"
)

var cmdData struct {
	OutputFormat string
	PrintValues  bool
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "get [revision]",
		Short:                 "Get information about a deployed release",
		DisableFlagsInUseLine: true,
		Args:                  cobra.MaximumNArgs(1),
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

			var revision int
			if len(args) > 0 {
				revision, err = strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("parse revision: %w", err)
				}
			}

			ctx = action.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultReleaseGetLogLevel), action.SetupLoggingOptions{ColorMode: lo.FromPtr(commonCmdData.LogColorMode)})

			if _, err = action.ReleaseGet(ctx, releaseName, releaseNamespace, action.ReleaseGetOptions{
				KubeConnectionOptions:       commonCmdData.KubeConnectionOptions,
				NetworkParallelism:          commonCmdData.NetworkParallelism,
				OutputFormat:                cmdData.OutputFormat,
				PrintValues:                 cmdData.PrintValues,
				ReleaseStorageDriver:        commonCmdData.ReleaseStorageDriver,
				ReleaseStorageSQLConnection: commonCmdData.ReleaseStorageSQLConnection,
				Revision:                    revision,
				TempDirPath:                 lo.FromPtr(commonCmdData.TmpDir),
			}); err != nil {
				return fmt.Errorf("release get: %w", err)
			}

			return nil
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	lo.Must0(common.SetupKubeConnectionFlags(&commonCmdData, cmd))
	common.SetupRelease(&commonCmdData, cmd, false)
	common.SetupNamespace(&commonCmdData, cmd, false)
	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupReleaseStorageDriver(&commonCmdData, cmd)
	common.SetupReleaseStorageSQLConnection(&commonCmdData, cmd)
	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.OutputFormat, "output-format", "", cmp.Or(os.Getenv("WERF_OUTPUT_FORMAT"), action.DefaultReleaseGetOutputFormat), "Output format. Options: yaml, json (default $WERF_OUTPUT_FORMAT or \""+action.DefaultReleaseGetOutputFormat+"\")")
	cmd.Flags().BoolVarP(&cmdData.PrintValues, "print-values", "", util.GetBoolEnvironmentDefaultFalse("WERF_PRINT_VALUES"), "Include computed values in output (default $WERF_PRINT_VALUES or false)")

	return cmd
}
