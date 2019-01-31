package purge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/common/docker_authorizer"
	"github.com/flant/werf/pkg/cleanup"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	DryRun bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "purge",
		DisableFlagsInUseLine: true,
		Short:                 "Purge project stages from stages storage",
		Long:                  common.GetLongCommandDescription("Purge project stages from stages storage"),
		RunE: func(cmd *cobra.Command, args []string) error {
			common.LogVersion()

			err := runPurge()
			if err != nil {
				return fmt.Errorf("stages purge failed: %s", err)
			}

			return nil
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupStagesRepo(&CommonCmdData, cmd)

	cmd.Flags().BoolVarP(&CmdData.DryRun, "dry-run", "", false, "Indicate what the command would do without actually doing that")

	return cmd
}

func runPurge() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}
	common.LogProjectDir(projectDir)

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("cannot parse werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	_, err = common.GetStagesRepo(&CommonCmdData)
	if err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	commonProjectOptions := cleanup.CommonProjectOptions{
		ProjectName:   projectName,
		CommonOptions: cleanup.CommonOptions{DryRun: CmdData.DryRun},
	}

	if err := cleanup.StagesPurge(commonProjectOptions); err != nil {
		return err
	}

	return nil
}
