package cleanup

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/cleaning"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "cleanup",
		DisableFlagsInUseLine: true,
		Short:                 "Cleanup project images",
		Long: common.GetLongCommandDescription(`Safely cleanup unused project images.

The command works according to special rules called cleanup policies, which the user defines in werf.yaml (https://werf.io/documentation/reference/werf_yaml.html#configuring-cleanup-policies).

It is safe to run this command periodically (daily is enough) by automated cleanup job in parallel with other werf commands such as build, converge and host cleanup.`),
		Example: `  $ werf cleanup --repo registry.mydomain.com/myproject/werf`,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer global_warnings.PrintGlobalWarnings(common.BackgroundContext())

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runCleanup()
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismInspectorOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupStagesStorageOptions(&commonCmdData, cmd)
	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultCleanupParallelTasksLimit)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and delete images from the specified repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupScanContextNamespaceOnly(&commonCmdData, cmd)
	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)
	common.SetupWithoutKube(&commonCmdData, cmd)
	common.SetupKeepStagesBuiltWithinLastNHours(&commonCmdData, cmd)

	return cmd
}

func runCleanup() error {
	tmp_manager.AutoGCEnabled = true
	ctx := common.BackgroundContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := common.InitGiterminismInspector(&commonCmdData); err != nil {
		return err
	}

	if err := git_repo.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
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

	common.SetupOndemandKubeInitializer(*commonCmdData.KubeContext, *commonCmdData.KubeConfig, *commonCmdData.KubeConfigBase64)
	if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&commonCmdData, projectDir)

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	localGitRepo, err := common.OpenLocalGitRepo(projectDir)
	if err != nil {
		return fmt.Errorf("unable to open local repo %s: %s", projectDir, err)
	}

	werfConfig, err := common.GetRequiredWerfConfig(ctx, projectDir, &commonCmdData, localGitRepo, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	if !werfConfig.Meta.GitWorktree.GetForceShallowClone() && !werfConfig.Meta.GitWorktree.GetAllowFetchingOriginBranchesAndTags() {
		isShallow, err := localGitRepo.IsShallowClone()
		if err != nil {
			return fmt.Errorf("check shallow clone failed: %s", err)
		}

		if isShallow {
			logboek.Warn().LogLn("Git shallow clone should not be used with images cleanup commands due to incompleteness of the repository history that is extremely essential for proper work.")
			logboek.Warn().LogLn("It is recommended to enable automatic fetch of origin git branches and tags during cleanup process with the gitWorktree.allowFetchOriginBranchesAndTags=true werf.yaml directive (which is enabled by default, http://werf.io/documentation/reference/werf_yaml.html#git-worktree).")
			logboek.Warn().LogLn("If you still want to use shallow clone, add gitWorktree.forceShallowClone=true directive into werf.yaml (http://werf.io/documentation/reference/werf_yaml.html#git-worktree).")

			return fmt.Errorf("git shallow clone is not allowed")
		}
	}

	if werfConfig.Meta.GitWorktree.GetAllowFetchingOriginBranchesAndTags() {
		if err := localGitRepo.SyncWithOrigin(ctx); err != nil {
			return fmt.Errorf("synchronization failed: %s", err)
		}
	}

	projectName := werfConfig.Meta.Project

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

	kubernetesContextClients, err := common.GetKubernetesContextClients(&commonCmdData)
	if err != nil {
		return fmt.Errorf("unable to get Kubernetes clusters connections: %s", err)
	}

	cleanupOptions := cleaning.CleanupOptions{
		ImageNameList:                           imagesNames,
		LocalGit:                                &localGitRepo,
		KubernetesContextClients:                kubernetesContextClients,
		KubernetesNamespaceRestrictionByContext: common.GetKubernetesNamespaceRestrictionByContext(&commonCmdData, kubernetesContextClients),
		WithoutKube:                             *commonCmdData.WithoutKube,
		GitHistoryBasedCleanupOptions:           werfConfig.Meta.Cleanup,
		KeepStagesBuiltWithinLastNHours:         *commonCmdData.KeepStagesBuiltWithinLastNHours,
		DryRun:                                  *commonCmdData.DryRun,
	}

	logboek.LogOptionalLn()
	if err := cleaning.Cleanup(ctx, projectName, storageManager, storageLockManager, cleanupOptions); err != nil {
		return err
	}

	return nil
}
