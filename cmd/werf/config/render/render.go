package render

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_inspector"
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

			if err := giterminism_inspector.Init(giterminism_inspector.InspectionOptions{LooseGiterminism: *commonCmdData.LooseGiterminism, NonStrict: *commonCmdData.NonStrictGiterminismInspection}); err != nil {
				return err
			}

			if err := git_repo.Init(); err != nil {
				return err
			}

			projectDir, err := common.GetProjectDir(&commonCmdData)
			if err != nil {
				return fmt.Errorf("getting project dir failed: %s", err)
			}

			localGitRepo, err := git_repo.OpenLocalRepo("own", projectDir)
			if err != nil {
				return fmt.Errorf("unable to open local repo %s: %s", projectDir, err)
			}

			configOpts := common.GetWerfConfigOptions(&commonCmdData, false)

			// TODO disable logboek only for this action
			werfConfigPath, err := common.GetWerfConfigPath(projectDir, *commonCmdData.ConfigPath, true, localGitRepo, configOpts)
			if err != nil {
				return err
			}

			werfConfigTemplatesDir := common.GetWerfConfigTemplatesDir(projectDir, &commonCmdData)

			return config.RenderWerfConfig(common.BackgroundContext(), projectDir, werfConfigPath, werfConfigTemplatesDir, args, localGitRepo, configOpts)
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupLooseGiterminism(&commonCmdData, cmd)
	common.SetupNonStrictGiterminismInspection(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupDev(&commonCmdData, cmd)

	return cmd
}
