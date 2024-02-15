package cleanup

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/cleaning"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

var cmdData struct {
	ScanContextOnly string
}

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "cleanup",
		DisableFlagsInUseLine: true,
		Short:                 "Cleanup project images in the container registry",
		Long:                  common.GetLongCommandDescription(GetCleanupDocs().Long),
		Example:               `  $ werf cleanup --repo registry.mydomain.com/myproject/werf`,
		Annotations: map[string]string{
			common.DocsLongMD: GetCleanupDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runCleanup(ctx)
			})
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})
	common.SetupFinalRepo(&commonCmdData, cmd)
	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultCleanupParallelTasksLimit)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and delete images from the specified repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupScanContextNamespaceOnly(&commonCmdData, cmd)
	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupWithoutKube(&commonCmdData, cmd)
	common.SetupKeepStagesBuiltWithinLastNHours(&commonCmdData, cmd)

	common.SetupDisableAutoHostCleanup(&commonCmdData, cmd)
	common.SetupAllowedDockerStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedDockerStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupDockerServerStoragePath(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)

	// aliases, but only WERF_SCAN_ONLY_CONTEXT env var is supported
	cmd.PersistentFlags().StringVarP(&cmdData.ScanContextOnly, "scan-context-only", "", os.Getenv("WERF_SCAN_CONTEXT_ONLY"), "Scan for used images only in the specified kube context, scan all contexts from kube config otherwise (default false or $WERF_SCAN_CONTEXT_ONLY)")
	cmd.PersistentFlags().StringVarP(&cmdData.ScanContextOnly, "kube-context", "", os.Getenv("WERF_SCAN_CONTEXT_ONLY"), "Scan for used images only in the specified kube context, scan all contexts from kube config otherwise (default false or $WERF_SCAN_CONTEXT_ONLY)")

	return cmd
}

func runCleanup(ctx context.Context) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	containerBackend, processCtx, err := common.InitProcessContainerBackend(ctx, &commonCmdData)
	if err != nil {
		return err
	}
	ctx = processCtx

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := lrumeta.Init(); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
		return err
	}

	defer func() {
		if err := common.RunAutoHostCleanup(ctx, &commonCmdData, containerBackend); err != nil {
			logboek.Context(ctx).Error().LogF("Auto host cleanup failed: %s\n", err)
		}
	}()

	common.SetupOndemandKubeInitializer(cmdData.ScanContextOnly, *commonCmdData.KubeConfig, *commonCmdData.KubeConfigBase64, *commonCmdData.KubeConfigPathMergeList)
	if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
		return err
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	if !werfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy {
		if !werfConfig.Meta.GitWorktree.GetForceShallowClone() && !werfConfig.Meta.GitWorktree.GetAllowFetchingOriginBranchesAndTags() {
			isShallow, err := giterminismManager.LocalGitRepo().IsShallowClone(ctx)
			if err != nil {
				return fmt.Errorf("check shallow clone failed: %w", err)
			}

			if isShallow {
				logboek.Warn().LogLn("Git shallow clone should not be used with images cleanup commands due to incompleteness of the repository history that is extremely essential for proper work.")
				logboek.Warn().LogLn("It is recommended to enable automatic fetch of origin git branches and tags during cleanup process with the gitWorktree.allowFetchOriginBranchesAndTags=true werf.yaml directive (which is enabled by default, http://werf.io/documentation/reference/werf_yaml.html#git-worktree).")
				logboek.Warn().LogLn("If you still want to use shallow clone, add gitWorktree.forceShallowClone=true directive into werf.yaml (http://werf.io/documentation/reference/werf_yaml.html#git-worktree).")

				return fmt.Errorf("git shallow clone is not allowed")
			}
		}

		if werfConfig.Meta.GitWorktree.GetAllowFetchingOriginBranchesAndTags() {
			if err := giterminismManager.LocalGitRepo().SyncWithOrigin(ctx); err != nil {
				return fmt.Errorf("synchronization failed: %w", err)
			}
		}
	}

	projectName := werfConfig.Meta.Project

	_, err = commonCmdData.Repo.GetAddress()
	if err != nil {
		logboek.Context(ctx).Default().LogLnDetails(`The "werf cleanup" command is only used to cleaning the container registry. In case you need to clean the runner or the localhost, use the commands of the "werf host" group.
It is worth noting that auto-cleaning is enabled by default, and manual use is usually not required (if not, we would appreciate feedback and creating an issue https://github.com/werf/werf/issues/new).`)

		return err
	}
	stagesStorage, err := common.GetStagesStorage(ctx, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}
	finalStagesStorage, err := common.GetOptionalFinalStagesStorage(ctx, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}

	synchronization, err := common.GetSynchronization(ctx, &commonCmdData, projectName, stagesStorage)
	if err != nil {
		return err
	}
	storageLockManager, err := common.GetStorageLockManager(ctx, synchronization)
	if err != nil {
		return err
	}
	secondaryStagesStorageList, err := common.GetSecondaryStagesStorageList(ctx, stagesStorage, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}
	cacheStagesStorageList, err := common.GetCacheStagesStorageList(ctx, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}

	storageManager := manager.NewStorageManager(projectName, stagesStorage, finalStagesStorage, secondaryStagesStorageList, cacheStagesStorageList, storageLockManager)

	if *commonCmdData.Parallel {
		storageManager.EnableParallel(int(*commonCmdData.ParallelTasksLimit))
	}

	imagesNames, err := common.GetManagedImagesNames(ctx, projectName, stagesStorage, werfConfig)
	if err != nil {
		return err
	}
	logboek.Debug().LogF("Managed images names: %v\n", imagesNames)

	var kubernetesContextClients []*kube.ContextClient
	var kubernetesNamespaceRestrictionByContext map[string]string
	if !(*commonCmdData.WithoutKube || werfConfig.Meta.Cleanup.DisableKubernetesBasedPolicy) {
		kubernetesContextClients, err = common.GetKubernetesContextClients(*commonCmdData.KubeConfig, *commonCmdData.KubeConfigBase64, *commonCmdData.KubeConfigPathMergeList, cmdData.ScanContextOnly)
		if err != nil {
			return fmt.Errorf("unable to get Kubernetes clusters connections: %w", err)
		}

		kubernetesNamespaceRestrictionByContext = common.GetKubernetesNamespaceRestrictionByContext(&commonCmdData, kubernetesContextClients)
	}

	cleanupOptions := cleaning.CleanupOptions{
		ImageNameList:                           imagesNames,
		LocalGit:                                giterminismManager.LocalGitRepo().(*git_repo.Local),
		KubernetesContextClients:                kubernetesContextClients,
		KubernetesNamespaceRestrictionByContext: kubernetesNamespaceRestrictionByContext,
		WithoutKube:                             *commonCmdData.WithoutKube,
		ConfigMetaCleanup:                       werfConfig.Meta.Cleanup,
		KeepStagesBuiltWithinLastNHours:         *commonCmdData.KeepStagesBuiltWithinLastNHours,
		DryRun:                                  *commonCmdData.DryRun,
	}

	logboek.LogOptionalLn()
	return cleaning.Cleanup(ctx, projectName, storageManager, cleanupOptions)
}
