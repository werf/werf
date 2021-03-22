package render

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "render [IMAGE_NAME...]",
		DisableFlagsInUseLine: true,
		Short:                 "Render werf.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := common.BackgroundContext()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
				return fmt.Errorf("initialization error: %s", err)
			}

			gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
			if err != nil {
				return fmt.Errorf("error getting host git data manager: %s", err)
			}

			if err := git_repo.Init(gitDataManager); err != nil {
				return err
			}

			if err := true_git.Init(true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
				return err
			}

			giterminismManager, err := common.GetGiterminismManager(&commonCmdData)
			if err != nil {
				return err
			}

			configOpts := common.GetWerfConfigOptions(&commonCmdData, false)

			customWerfConfigRelPath, err := common.GetCustomWerfConfigRelPath(giterminismManager, &commonCmdData)
			if err != nil {
				return err
			}

			customWerfConfigTemplatesDirRelPath, err := common.GetCustomWerfConfigTemplatesDirRelPath(giterminismManager, &commonCmdData)
			if err != nil {
				return err
			}

			return config.RenderWerfConfig(common.BackgroundContext(), customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath, args, giterminismManager, configOpts)
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}
