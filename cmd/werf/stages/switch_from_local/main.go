package switch_from_local

import (
	"fmt"
	"strings"

	stages_common "github.com/werf/werf/cmd/werf/stages/common"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/pkg/storage"

	"github.com/spf13/cobra"
	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/stages_manager"
	"github.com/werf/werf/pkg/werf"
)

var cmdData stages_common.SyncCmdData
var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "switch-from-local",
		DisableFlagsInUseLine: true,
		Short:                 "Switch current project stages storage from :local to repo",
		Long:                  common.GetLongCommandDescription("Switch current project stages storage to another"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runSwitch()
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	stages_common.SetupFromStagesStorage(&commonCmdData, &cmdData, cmd)
	stages_common.SetupToStagesStorage(&commonCmdData, &cmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	return cmd
}

func runSwitch() error {
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

	fromStagesStorage, err := stages_common.NewFromStagesStorage(&commonCmdData, &cmdData, containerRuntime, storage.LocalStorageAddress)
	if err != nil {
		return err
	}
	if fromStagesStorage.Address() != storage.LocalStorageAddress {
		return fmt.Errorf("cannot switch from non-local stages storage, omit --from param or specify --from=%s", storage.LocalStorageAddress)
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

	stagesManager := stages_manager.NewStagesManager(projectName, storageLockManager, stagesStorageCache)

	if err := stagesManager.UseStagesStorage(fromStagesStorage); err != nil {
		return err
	}

	toStagesStorage, err := stages_common.NewToStagesStorage(&commonCmdData, &cmdData, containerRuntime)
	if err != nil {
		return err
	}
	if toStagesStorage.Address() == storage.LocalStorageAddress {
		return fmt.Errorf("cannot switch to local stages storage, specify repo address --to=REPO")
	}

	if err := stages_manager.SyncStages(projectName, fromStagesStorage, toStagesStorage, storageLockManager, containerRuntime, stages_manager.SyncStagesOptions{}); err != nil {
		return err
	}

	if lock, err := storageLockManager.LockStagesAndImages(projectName, storage.LockStagesAndImagesOptions{}); err != nil {
		return err
	} else {
		defer storageLockManager.Unlock(lock)
	}

	if err := stagesManager.SetStagesSwitchFromLocalBlock(toStagesStorage); err != nil {
		return err
	}

	return stages_manager.SyncStages(projectName, fromStagesStorage, toStagesStorage, storageLockManager, containerRuntime, stages_manager.SyncStagesOptions{RemoveSource: true, CleanupLocalCache: true, WithoutLock: true})
}
