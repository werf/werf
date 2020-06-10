package sync

import (
	"fmt"
	"strings"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/werf/werf/pkg/stages_manager"

	"github.com/werf/werf/pkg/image"

	"github.com/flant/logboek"
	"github.com/werf/werf/cmd/werf/common"
	stages_common "github.com/werf/werf/cmd/werf/stages/common"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/werf"
	"github.com/spf13/cobra"
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
	common.SetupKubeContext(&commonCmdData, cmd)

	return cmd
}

func runSync() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(&commonCmdData); err != nil {
		return err
	}

	if err := docker.Init(*commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&commonCmdData, projectDir)

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, &commonCmdData, true)
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

	synchronization, err := common.GetSynchronization(&commonCmdData, fromStagesStorage.Address())
	if err != nil {
		return err
	}
	if strings.HasPrefix(synchronization, "kubernetes://") {
		if err := kube.Init(kube.InitOptions{KubeContext: *commonCmdData.KubeContext, KubeConfig: *commonCmdData.KubeConfig}); err != nil {
			return fmt.Errorf("cannot initialize kube: %s", err)
		}
	}
	stagesStorageCache, err := common.GetStagesStorageCache(synchronization)
	if err != nil {
		return err
	}
	storageLockManager, err := common.GetStorageLockManager(synchronization)
	if err != nil {
		return err
	}
	_ = stagesStorageCache

	return stages_manager.SyncStages(projectName, fromStagesStorage, toStagesStorage, storageLockManager, containerRuntime, stages_manager.SyncStagesOptions{RemoveSource: *cmdData.RemoveSource, CleanupLocalCache: *cmdData.CleanupLocalCache})
}
