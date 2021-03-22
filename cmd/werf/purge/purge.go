package purge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/cleaning"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	Force bool
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "purge",
		DisableFlagsInUseLine: true,
		Short:                 "Purge all project images from images repo and stages from stages storage",
		Long: common.GetLongCommandDescription(`Purge all project images from images repo and stages from stages storage.

First step is 'werf images purge', which will delete all project images from images repo. Second step is 'werf stages purge', which will delete all stages from stages storage.

WARNING: Do not run this command during any other werf command is working on the host machine. This command is supposed to be run manually. Images from images repo, that are being used in Kubernetes cluster will also be deleted.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			defer werf.PrintGlobalWarnings(common.BackgroundContext())

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runPurge()
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupStagesStorageOptions(&commonCmdData, cmd)
	common.SetupImagesRepoOptions(&commonCmdData, cmd)
	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultCleanupParallelTasksLimit)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to delete images from the specified stages storage and images repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)
	cmd.Flags().BoolVarP(&cmdData.Force, "force", "", false, common.CleaningCommandsForceOptionDescription)

	return cmd
}

func runPurge() error {
	tmp_manager.AutoGCEnabled = true
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

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&commonCmdData, projectDir)

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

	werfConfig, err := common.GetRequiredWerfConfig(ctx, projectDir, &commonCmdData, true)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	logboek.LogOptionalLn()

	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO

	stagesStorage, err := common.GetStagesStorage(containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}

	synchronization, err := common.GetSynchronization(ctx, &commonCmdData, projectName, stagesStorage)
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

	storageManager := manager.NewStorageManager(projectName, storageLockManager, stagesStorageCache)
	if err := storageManager.UseStagesStorage(ctx, stagesStorage); err != nil {
		return err
	}

	if stagesStorage.Address() != storage.LocalStorageAddress && *commonCmdData.Parallel {
		storageManager.StagesStorageManager.EnableParallel(int(*commonCmdData.ParallelTasksLimit))
	}

	imagesRepo, err := common.GetImagesRepo(ctx, projectName, &commonCmdData)
	if err != nil {
		return err
	}

	storageManager.SetImageRepo(imagesRepo)
	if *commonCmdData.Parallel {
		storageManager.ImagesRepoManager.EnableParallel(int(*commonCmdData.ParallelTasksLimit))
	}

	imagesNames, err := common.GetManagedImagesNames(ctx, projectName, stagesStorage, werfConfig)
	if err != nil {
		return err
	}
	logboek.Debug().LogF("Managed images names: %v\n", imagesNames)

	purgeOptions := cleaning.PurgeOptions{
		ImagesPurgeOptions: cleaning.ImagesPurgeOptions{
			ImageNameList: imagesNames,
			DryRun:        *commonCmdData.DryRun,
		},
		StagesPurgeOptions: cleaning.StagesPurgeOptions{
			RmContainersThatUseWerfImages: cmdData.Force,
			DryRun:                        *commonCmdData.DryRun,
		},
	}

	logboek.LogOptionalLn()
	if err := cleaning.Purge(ctx, projectName, storageManager, storageLockManager, purgeOptions); err != nil {
		return err
	}

	return nil
}
