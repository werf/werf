package sync

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	stages_common "github.com/werf/werf/cmd/werf/stages/common"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/werf"
)

var cmdData stages_common.SyncCmdData
var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "sync",
		DisableFlagsInUseLine: true,
		Short:                 "Sync project stages from one stages storage to another",
		Long:                  common.GetLongCommandDescription("Sync project stages from one stages storage to another"),
		RunE: func(cmd *cobra.Command, args []string) error {
			defer werf.PrintGlobalWarnings(common.BackgroundContext())

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runSync()
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and delete images from the specified stages storages")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	stages_common.SetupFromStagesStorage(&commonCmdData, &cmdData, cmd)
	stages_common.SetupToStagesStorage(&commonCmdData, &cmdData, cmd)
	stages_common.SetupRemoveSource(&cmdData, cmd)
	stages_common.SetupCleanupLocalCache(&cmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	return cmd
}

func runSync() error {
	ctx := common.BackgroundContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := lrumeta.Init(); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(&commonCmdData); err != nil {
		return err
	}

	if err := docker.Init(ctx, *commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	ctxWithDockerCli, err := docker.NewContext(ctx)
	if err != nil {
		return err
	}
	ctx = ctxWithDockerCli

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&commonCmdData, projectDir)

	werfConfig, err := common.GetRequiredWerfConfig(ctx, projectDir, &commonCmdData, true)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	logboek.LogOptionalLn()

	projectName := werfConfig.Meta.Project

	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO

	fromStagesStorage, err := stages_common.NewFromStagesStorage(&commonCmdData, &cmdData, containerRuntime, "")
	if err != nil {
		return err
	}

	toStagesStorage, err := stages_common.NewToStagesStorage(&commonCmdData, &cmdData, containerRuntime)
	if err != nil {
		return err
	}

	synchronization, err := common.GetSynchronization(ctx, &commonCmdData, projectName, fromStagesStorage)
	if err != nil {
		return err
	}
	stagesStorageCache, err := common.GetStagesStorageCache(synchronization)
	if err != nil {
		return err
	}
	storageLockManager, err := common.GetStorageLockManager(ctx, synchronization)
	if err != nil {
		return err
	}
	_ = stagesStorageCache

	return manager.SyncStages(ctx, projectName, fromStagesStorage, toStagesStorage, storageLockManager, containerRuntime, manager.SyncStagesOptions{RemoveSource: *cmdData.RemoveSource, CleanupLocalCache: *cmdData.CleanupLocalCache})
}
