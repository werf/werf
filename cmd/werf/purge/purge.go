package purge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/cleaning"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_inspector"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var cmdData struct {
	Force bool
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "purge",
		DisableFlagsInUseLine: true,
		Short:                 "Purge all project images",
		Long: common.GetLongCommandDescription(`Purge all project images.

WARNING: Do not run this command during any other werf command is working on the host machine. This command is supposed to be run manually. Images from repo, that are being used in Kubernetes cluster will also be deleted.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			defer global_warnings.PrintGlobalWarnings(common.BackgroundContext())

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
	common.SetupDisableGiterminism(&commonCmdData, cmd)
	common.SetupNonStrictGiterminismInspection(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupStagesStorageOptions(&commonCmdData, cmd)
	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultCleanupParallelTasksLimit)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to delete images from the specified repo")
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

	if err := giterminism_inspector.Init(giterminism_inspector.InspectionOptions{DisableGiterminism: *commonCmdData.DisableGiterminism, NonStrict: *commonCmdData.NonStrictGiterminismInspection}); err != nil {
		return err
	}

	if err := git_repo.Init(); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
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

	localGitRepo, err := git_repo.OpenLocalRepo("own", projectDir)
	if err != nil {
		return fmt.Errorf("unable to open local repo %s: %s", projectDir, err)
	}

	werfConfig, err := common.GetRequiredWerfConfig(ctx, projectDir, &commonCmdData, localGitRepo, config.WerfConfigOptions{LogRenderedFilePath: true, DisableGiterminism: *commonCmdData.DisableGiterminism, Env: *commonCmdData.Environment})
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	logboek.LogOptionalLn()

	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO

	stagesStorageAddress := common.GetOptionalStagesStorageAddress(&commonCmdData)
	stagesStorage, err := common.GetStagesStorage(stagesStorageAddress, containerRuntime, &commonCmdData)
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
	secondaryStagesStorageList, err := common.GetSecondaryStagesStorageList(stagesStorage, containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}

	storageManager := manager.NewStorageManager(projectName, stagesStorage, secondaryStagesStorageList, storageLockManager, stagesStorageCache)

	if stagesStorage.Address() != storage.LocalStorageAddress && *commonCmdData.Parallel {
		storageManager.StagesStorageManager.EnableParallel(int(*commonCmdData.ParallelTasksLimit))
	}

	imagesNames, err := common.GetManagedImagesNames(ctx, projectName, stagesStorage, werfConfig)
	if err != nil {
		return err
	}
	logboek.Debug().LogF("Managed images names: %v\n", imagesNames)

	purgeOptions := cleaning.PurgeOptions{
		RmContainersThatUseWerfImages: cmdData.Force,
		DryRun:                        *commonCmdData.DryRun,
	}

	logboek.LogOptionalLn()
	if err := cleaning.Purge(ctx, projectName, storageManager, storageLockManager, purgeOptions); err != nil {
		return err
	}

	return nil
}
