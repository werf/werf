package render

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/werf"
)

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "render [IMAGE_NAME...]",
		DisableFlagsInUseLine: true,
		Short:                 "Render werf.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
				return fmt.Errorf("initialization error: %s", err)
			}

			tmp_manager.AutoGCEnabled = false

			projectDir, err := common.GetProjectDir(&CommonCmdData)
			if err != nil {
				return fmt.Errorf("getting project dir failed: %s", err)
			}

			werfConfigPath, err := common.GetWerfConfigPath(projectDir)
			if err != nil {
				return err
			}

			return config.RenderWerfConfig(werfConfigPath, args)
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	return cmd
}
