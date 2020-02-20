package list

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	Force bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "purge [PROJECT_NAME ...]",
		Short:                 "Purge project stages from local stages storage",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&CommonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			if len(args) == 0 {
				common.PrintHelp(cmd)
				return fmt.Errorf("accepts position arguments, received %d", len(args))
			}

			return common.LogRunningTime(func() error {
				return run(args...)
			})
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	common.SetupStagesStorage(&CommonCmdData, cmd)
	common.SetupStagesStorageLock(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command needs granted permissions to read, pull and delete images from the specified stages storage")

	common.SetupLogOptions(&CommonCmdData, cmd)

	common.SetupDryRun(&CommonCmdData, cmd)
	cmd.Flags().BoolVarP(&CmdData.Force, "force", "", false, common.CleaningCommandsForceOptionDescription)

	return cmd
}

func run(projectNames ...string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
		return err
	}

	if err := docker_registry.Init(docker_registry.Options{InsecureRegistry: false, SkipTlsVerifyRegistry: false}); err != nil {
		return err
	}

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	logboek.LogOptionalLn()

	for _, projectName := range projectNames {
		logProcessOptions := logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()}
		if err := logboek.Default.LogProcess("Project "+projectName, logProcessOptions, func() error {
			stagesPurgeOptions := cleaning.StagesPurgeOptions{
				ProjectName:                   projectName,
				RmContainersThatUseWerfImages: CmdData.Force,
				DryRun:                        *CommonCmdData.DryRun,
			}

			if err := cleaning.StagesPurge(stagesPurgeOptions); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}
