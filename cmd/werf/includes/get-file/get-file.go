package getfile

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
		Use:   "get-file [FILE_NAME]",
		Short: "Display file content that will be used by werf according to the includes overlay rules.",
		Long:  "Display file content that will be used by werf according to the includes overlay rules.",
		Example: `
  # Display file content
  $ werf includes get-file werf.yaml
`,
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

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			gm, err := common.GetGiterminismManager(ctx, &commonCmdData)
			if err != nil {
				return err
			}

			content, err := gm.FileManager.ConfigGoTemplateFilesGet(ctx, args[0])
			if err != nil {
				return fmt.Errorf("unable to get file: %w", err)
			}

			fmt.Print(string(content))

			return nil
		},
	}))

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupLogOptionsDefaultQuiet(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	commonCmdData.SetupPlatform(cmd)
	common.SetupFollow(&commonCmdData, cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	return cmd
}
