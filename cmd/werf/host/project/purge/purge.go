package list

import (
	"fmt"
	"path/filepath"

	"github.com/flant/werf/pkg/stages_manager"
	"github.com/flant/werf/pkg/storage"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/werf"
)

var cmdData struct {
	Force bool
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "purge [PROJECT_NAME ...]",
		Short:                 "Purge project stages from local stages storage",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
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

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupStagesStorageOptions(&commonCmdData, cmd) // TODO: host project purge command should process only :local stages storage
	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and delete images from the specified stages storage")

	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)
	cmd.Flags().BoolVarP(&cmdData.Force, "force", "", false, common.CleaningCommandsForceOptionDescription)

	return cmd
}

func run(projectNames ...string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
		return err
	}

	if err := docker.Init(*commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO

	stagesStorage, err := common.GetStagesStorage(containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}

	stagesStorageCache := common.GetStagesStorageCache()
	storageLockManager := &storage.FileLockManager{}

	logboek.LogOptionalLn()

	for _, projectName := range projectNames {
		stagesManager := stages_manager.NewStagesManager(projectName, storageLockManager, stagesStorage, stagesStorageCache)

		logProcessOptions := logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()}
		if err := logboek.Default.LogProcess("Project "+projectName, logProcessOptions, func() error {
			stagesPurgeOptions := cleaning.StagesPurgeOptions{
				RmContainersThatUseWerfImages: cmdData.Force,
				DryRun:                        *commonCmdData.DryRun,
			}

			return cleaning.StagesPurge(projectName, stagesManager, stagesPurgeOptions)
		}); err != nil {
			return err
		}
	}

	return nil
}
