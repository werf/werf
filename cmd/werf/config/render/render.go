package render

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "render [IMAGE_NAME...]",
		DisableFlagsInUseLine: true,
		Short:                 "Render werf.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
				return fmt.Errorf("initialization error: %s", err)
			}

			projectDir, err := common.GetProjectDir(&commonCmdData)
			if err != nil {
				return fmt.Errorf("getting project dir failed: %s", err)
			}

			localGitRepo, err := git_repo.OpenLocalRepo("own", projectDir)
			if err != nil {
				return fmt.Errorf("unable to open local repo %s: %s", projectDir, err)
			}

			configOpts := config.WerfConfigOptions{DisableDeterminism: *commonCmdData.DisableDeterminism}

			// TODO disable logboek only for this action
			werfConfigPath, err := common.GetWerfConfigPath(projectDir, &commonCmdData, true, localGitRepo, configOpts)
			if err != nil {
				return err
			}

			werfConfigTemplatesDir := common.GetWerfConfigTemplatesDir(projectDir, &commonCmdData)

			return config.RenderWerfConfig(common.BackgroundContext(), werfConfigPath, werfConfigTemplatesDir, args, localGitRepo, configOpts)
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupDisableDeterminism(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}
