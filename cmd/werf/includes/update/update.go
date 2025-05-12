package update

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/true_git"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, common.SetCommandContext(ctx, &cobra.Command{
		Use:   "update",
		Short: "Update werf-includes.lock file",
		Long:  "Update werf-includes.lock file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
				Cmd:                &commonCmdData,
				InitWerf:           true,
				InitGitDataManager: true,
				InitTrueGitWithOptions: &common.InitTrueGitOptions{
					Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
				},
			})
			if err != nil {
				return fmt.Errorf("component init error: %w", err)
			}

			_, err = common.GetGiterminismManager(ctx, &commonCmdData)
			if err != nil {
				return err
			}

			return nil
		},
	}))

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	commonCmdData.SetupPlatform(cmd)
	common.SetupFollow(&commonCmdData, cmd)

	commonCmdData.SetupCreateIncludesLockFile(true)
	commonCmdData.SetupUseIncludesLatestVersions(cmd)

	return cmd
}
