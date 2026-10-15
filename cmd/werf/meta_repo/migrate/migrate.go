package migrate

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/synchronization/lock_manager"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
)

var (
	commonCmdData      common.CmdData
	removeSource       bool
	parallelTasksLimit int
	migrationTimeout   time.Duration
)

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "migrate",
		DisableFlagsInUseLine: true,
		Short:                 "Move existing project metadata from --from into --to and enable the safeguard.",
		Long:                  "Copy the project's managed-images, image-metadata-by-commit, custom-tag metadata and last-cleanup record from --from into --to, then record the meta-repo marker in --from so subsequent runs are forced to use the same --meta-repo. Copy is verified before any source record is removed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if parallelTasksLimit < 1 {
				return fmt.Errorf("--parallel-tasks-limit must be positive")
			}
			if migrationTimeout < 0 {
				return fmt.Errorf("--timeout must not be negative")
			}

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return run(ctx)
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigRenderPath(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	setupFromRepo(&commonCmdData, cmd)
	commonCmdData.MetaRepo = common.NewRepoData("to", common.RepoDataOptions{OnlyAddress: true})
	commonCmdData.MetaRepo.SetupAddressForRepoData(cmd, "to", []string{"WERF_META_REPO_MIGRATE_TO"})

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and write images to --from and --to")
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	commonCmdData.SetupPlatform(cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	cmd.Flags().BoolVar(&removeSource, "remove-source", util.GetBoolEnvironmentDefaultTrue("WERF_REMOVE_SOURCE"), "Delete the original metadata records from --from after they are verified present in --to (default $WERF_REMOVE_SOURCE or true if not specified)")

	lo.Must0(common.SetupMinimalKubeConnectionFlags(&commonCmdData, cmd))

	parallelDefault, err := util.GetInt64EnvVar("WERF_PARALLEL_TASKS_LIMIT")
	if err != nil {
		panic(err)
	}
	cmd.Flags().IntVar(&parallelTasksLimit, "parallel-tasks-limit", int(lo.FromPtrOr(parallelDefault, 5)), "Maximum number of concurrent metadata operations, must be positive (default $WERF_PARALLEL_TASKS_LIMIT or 5)")
	timeoutDefault, err := util.GetDurationEnvVar("WERF_META_REPO_MIGRATE_TIMEOUT")
	if err != nil {
		panic(err)
	}
	cmd.Flags().DurationVar(&migrationTimeout, "timeout", timeoutDefault, "Request cancellation of registry migration after this duration, starting after lock acquisition (e.g. 30m, 2h); 0 disables the limit (default $WERF_META_REPO_MIGRATE_TIMEOUT). Rerun the command to resume after interruption")

	return cmd
}

// Source deletion needs registry credentials scoped separately from other copy commands.
func setupFromRepo(cmdData *common.CmdData, cmd *cobra.Command) {
	cmdData.Repo = common.NewRepoData("from", common.RepoDataOptions{})
	cmdData.Repo.SetupAddressForRepoData(cmd, "from", []string{"WERF_META_REPO_MIGRATE_FROM"})
	cmdData.Repo.SetupContainerRegistryForRepoData(cmd, "from-container-registry", []string{"WERF_META_REPO_MIGRATE_FROM_CONTAINER_REGISTRY"})
	cmdData.Repo.SetupDockerHubUsernameForRepoData(cmd, "from-docker-hub-username", []string{"WERF_META_REPO_MIGRATE_FROM_DOCKER_HUB_USERNAME"})
	cmdData.Repo.SetupDockerHubPasswordForRepoData(cmd, "from-docker-hub-password", []string{"WERF_META_REPO_MIGRATE_FROM_DOCKER_HUB_PASSWORD"})
	cmdData.Repo.SetupDockerHubTokenForRepoData(cmd, "from-docker-hub-token", []string{"WERF_META_REPO_MIGRATE_FROM_DOCKER_HUB_TOKEN"})
	cmdData.Repo.SetupGithubTokenForRepoData(cmd, "from-github-token", []string{"WERF_META_REPO_MIGRATE_FROM_GITHUB_TOKEN"})
	cmdData.Repo.SetupHarborUsernameForRepoData(cmd, "from-harbor-username", []string{"WERF_META_REPO_MIGRATE_FROM_HARBOR_USERNAME"})
	cmdData.Repo.SetupHarborPasswordForRepoData(cmd, "from-harbor-password", []string{"WERF_META_REPO_MIGRATE_FROM_HARBOR_PASSWORD"})
	cmdData.Repo.SetupQuayTokenForRepoData(cmd, "from-quay-token", []string{"WERF_META_REPO_MIGRATE_FROM_QUAY_TOKEN"})
}

func run(ctx context.Context) error {
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
		InitDockerRegistry:           true,
		InitProcessContainerBackend:  true,
		InitWerf:                     true,
		InitGitDataManager:           true,
		InitManifestCache:            true,
		InitLRUImagesCache:           true,
		SetupOndemandKubeInitializer: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	containerBackend := commonManager.ContainerBackend()

	if _, err = tmp_manager.CreateProjectDir(ctx); err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	_, werfConfig, err := common.GetOptionalWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}
	if werfConfig == nil {
		return fmt.Errorf("run command in the project directory with werf.yaml")
	}
	projectName := werfConfig.Meta.Project

	storageManager, err := common.NewStorageManager(ctx, &common.NewStorageManagerConfig{
		ProjectName:                    projectName,
		SkipMetaRepoSafeguard:          true,
		ContainerBackend:               containerBackend,
		CmdData:                        &commonCmdData,
		CleanupDisabled:                werfConfig.Meta.Cleanup.DisableCleanup,
		GitHistoryBasedCleanupDisabled: werfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy,
	})
	if err != nil {
		return fmt.Errorf("unable to init storage manager: %w", err)
	}

	src, ok := storageManager.GetStagesStorage().(*storage.RepoStagesStorage)
	if !ok {
		return fmt.Errorf("--from must be a container registry address")
	}
	dst, ok := storageManager.GetMetaStorage(ctx).(*storage.RepoStagesStorage)
	if !ok || dst == src {
		return fmt.Errorf("--to must be set to a container registry address different from --from")
	}

	if err := migrateWithLock(ctx, projectName, src, dst, storageManager.StorageLockManager, storage.MigrateMetaRepoOptions{RemoveSource: removeSource, ParallelTasksLimit: parallelTasksLimit}, migrationTimeout); err != nil {
		return fmt.Errorf("unable to migrate metadata for project %q: %w", projectName, err)
	}

	logboek.Context(ctx).Default().LogF("Metadata for project %q migrated to meta-repo %s; safeguard enabled in %s\n", projectName, dst.Address(), src.Address())
	return nil
}

func migrateWithLock(ctx context.Context, projectName string, src, dst *storage.RepoStagesStorage, locks lock_manager.Interface, opts storage.MigrateMetaRepoOptions, timeout time.Duration) error {
	source, err := storage.CanonicalRepoAddress(ctx, src.Address())
	if err != nil {
		return err
	}
	lockName := fmt.Sprintf("meta-repo-migrate-%x", sha256.Sum256([]byte(source)))
	var handle lock_manager.LockHandle
	if err := logboek.Context(ctx).Default().LogProcess("Waiting for metadata migration lock on %s", source).DoError(func() error {
		var err error
		handle, err = locks.LockStage(ctx, projectName, lockName)
		return err
	}); err != nil {
		return fmt.Errorf("lock metadata migration: %w", err)
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	err = storage.MigrateMetaRepo(ctx, projectName, src, dst, opts)
	unlockErr := locks.Unlock(context.WithoutCancel(ctx), handle)
	if unlockErr != nil {
		unlockErr = fmt.Errorf("unlock metadata migration: %w", unlockErr)
	}
	return errors.Join(err, unlockErr)
}
