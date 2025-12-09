package copy

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/build/stages"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/storage/manager"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

type copyCmdData struct {
	From string
	To   string
	All  bool
}

type copyOptions struct {
	From *ref.Addr
	To   *ref.Addr
	All  bool
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	var cmdData copyCmdData

	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "copy",
		Short:                 "Copy stages between container registry and archive",
		Long:                  common.GetLongCommandDescription(GetCopyDocs().Long),
		DisableFlagsInUseLine: true,
		Example: `  # Copy stages between container registries
  $ werf stages copy \
      --from index.docker.io/company/first-project \
      --to index.docker.io/company/second-project

  # Copy stages between container registry and archive
  $ werf stages copy \
      --from index.docker.io/company/project \
      --to archive:/path/to/archive.tar.gz

  # Copy stages between archive and container registry
  $ werf stages copy \
      --from archive:/path/to/archive.tar.gz \
      --to index.docker.io/company/project`,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
			common.DocsLongMD: GetCopyDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runCopy(ctx, cmdData) })
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigRenderPath(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)

	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})
	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	common.SetupFinalRepo(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repos")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupIntrospectAfterError(&commonCmdData, cmd)
	common.SetupIntrospectBeforeError(&commonCmdData, cmd)
	common.SetupIntrospectStage(&commonCmdData, cmd)

	common.SetupSaveBuildReport(&commonCmdData, cmd)
	common.SetupBuildReportPath(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupProjectName(&commonCmdData, cmd, false)
	common.SetupBackendStoragePath(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	commonCmdData.SetupFinalImagesOnly(cmd, false)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	commonCmdData.SetupSkipImageSpecStage(cmd)
	commonCmdData.SetupDebugTemplates(cmd)

	commonCmdData.SetupPlatform(cmd)

	setupCopyOptions(&cmdData, cmd)

	return cmd
}

func runCopy(ctx context.Context, cmdData copyCmdData) error {
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:                         &commonCmdData,
		InitWerf:                    true,
		InitGitDataManager:          true,
		InitDockerRegistry:          true,
		InitProcessContainerBackend: true,
		InitManifestCache:           true,
		InitLRUImagesCache:          true,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	opts, err := getCopyOptions(cmdData)
	if err != nil {
		return err
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}
	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	projectName := werfConfig.Meta.Project
	disableCleanup := werfConfig.Meta.Cleanup.DisableCleanup
	disableGitHistoryBasedPolicy := werfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy
	containerBackend := commonManager.ContainerBackend()

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}

	storageManager, dockerRegistry, err := initCommonCopyComponents(
		ctx,
		&common.NewStorageManagerConfig{
			ProjectName:                    projectName,
			ContainerBackend:               containerBackend,
			CmdData:                        &commonCmdData,
			CleanupDisabled:                disableCleanup,
			GitHistoryBasedCleanupDisabled: disableGitHistoryBasedPolicy,
		},
	)
	if err != nil {
		return fmt.Errorf("unable to init common components: %w", err)
	}

	var conveyorWithRetryWrapper *build.ConveyorWithRetryWrapper
	var buildOptions build.BuildOptions

	if !cmdData.All {
		conveyorWithRetryWrapper, buildOptions, err = initConveyorComponents(ctx, werfConfig, giterminismManager, projectTmpDir, containerBackend, storageManager)
		if err != nil {
			return fmt.Errorf("unable to init components: %w", err)
		}
		defer conveyorWithRetryWrapper.Terminate()
	}

	return logboek.Context(ctx).LogProcess("Copy stages").DoError(func() error {
		logboek.Context(ctx).LogFDetails("From: %s\n", cmdData.From)
		logboek.Context(ctx).LogFDetails("To: %s\n", cmdData.To)

		return stages.Copy(ctx, opts.From, opts.To, stages.CopyOptions{
			All:               cmdData.All,
			ProjectName:       werfConfig.Meta.Project,
			RegistryClient:    dockerRegistry,
			StorageManager:    storageManager,
			ConveyorWithRetry: conveyorWithRetryWrapper,
			BuildOptions:      buildOptions,
		})
	})
}

func setupCopyOptions(cmdData *copyCmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.From, "from", "", "", "Source address to copy stages from. Use archive:PATH for stage archive or [docker://]REPO for container registry.")
	cmd.Flags().StringVarP(&cmdData.To, "to", "", "", "Destination address to copy stages to. Use archive:PATH for stage archive or [docker://]REPO for container registry.")
	cmd.Flags().BoolVarP(&cmdData.All, "all", "", true, `Copy all project stages (default: true). Use --all=false to copy stages for current build only. Note: this flag is ignored when copying from archive to container registry.`)
}

func initCommonCopyComponents(ctx context.Context, managerConfig *common.NewStorageManagerConfig) (*manager.StorageManager, docker_registry.Interface, error) {
	storageManager, err := common.NewStorageManager(ctx, managerConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init storage manager: %w", err)
	}

	dockerRegistry, err := common.CreateDockerRegistry(ctx, *managerConfig.CmdData.Repo.Address, *commonCmdData.InsecureRegistry, *commonCmdData.SkipTlsVerifyRegistry)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create docker registry: %w", err)
	}

	return storageManager, dockerRegistry, nil
}

func initConveyorComponents(ctx context.Context, werfConfig *config.WerfConfig, giterminismManager giterminism_manager.Interface, projectTmpDir string, containerBackend container_backend.ContainerBackend, storageManager *manager.StorageManager) (*build.ConveyorWithRetryWrapper, build.BuildOptions, error) {
	imageToProcess, err := config.NewImagesToProcess(werfConfig, nil, *commonCmdData.FinalImagesOnly, false)
	if err != nil {
		return nil, build.BuildOptions{}, fmt.Errorf("unable to get images to process: %w", err)
	}

	buildOptions, err := common.GetBuildOptions(ctx, &commonCmdData, werfConfig, imageToProcess)
	if err != nil {
		return nil, build.BuildOptions{}, fmt.Errorf("unable to get build options: %w", err)
	}

	conveyorOptions, err := common.GetConveyorOptionsWithParallel(ctx, &commonCmdData, imageToProcess, buildOptions)
	if err != nil {
		return nil, build.BuildOptions{}, fmt.Errorf("unable to get conveyor options: %w", err)
	}

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, containerBackend, storageManager, storageManager.StorageLockManager, conveyorOptions)

	return conveyorWithRetry, buildOptions, nil
}

func getCopyOptions(cmdData copyCmdData) (copyOptions, error) {
	if err := validateRawCopyOptions(cmdData); err != nil {
		return copyOptions{}, err
	}

	fromAddr, err := ref.ParseAddr(cmdData.From)
	if err != nil {
		return copyOptions{}, fmt.Errorf("invalid from addr %q: %w", cmdData.From, err)
	}

	toAddr, err := ref.ParseAddr(cmdData.To)
	if err != nil {
		return copyOptions{}, fmt.Errorf("invalid to addr %q: %w", cmdData.To, err)
	}

	if fromAddr.RegistryAddress != nil {
		commonCmdData.Repo.Address = &fromAddr.RegistryAddress.Repo // FIXME выдумать что-нить симпатичнее
	} else if toAddr.RegistryAddress != nil {
		commonCmdData.Repo.Address = &toAddr.RegistryAddress.Repo
	} else {
		return copyOptions{}, fmt.Errorf("--from or --to address must be container registry address")
	}

	opts := copyOptions{
		From: fromAddr,
		To:   toAddr,
		All:  cmdData.All,
	}

	return opts, nil
}

func validateRawCopyOptions(cmdData copyCmdData) error {
	if cmdData.From == "" {
		return errors.New("--from=ADDRESS param required")
	}

	if cmdData.To == "" {
		return errors.New("--to=ADDRESS param required")
	}

	if cmdData.From == cmdData.To {
		return errors.New("--from=ADDRESS and --to=ADDRESS must be different")
	}

	return nil
}
