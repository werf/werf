package copy

import (
	"context"
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

var cmdData struct {
	From string
	To   string
	All  bool
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "copy",
		Short:                 "Copy stages between container registry and archive",
		Long:                  common.GetLongCommandDescription(GetCopyDocs().Long),
		DisableFlagsInUseLine: true,
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

			return common.LogRunningTime(func() error { return runCopy(ctx) })
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})
	common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repos")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupFinalRepo(&commonCmdData, cmd)
	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	commonCmdData.SetupFinalImagesOnly(cmd, false)
	commonCmdData.SetupPlatform(cmd)

	cmd.Flags().StringVarP(&cmdData.From, "from", "", "", "Source address to copy stages from. Use archive:PATH for stage archive or [docker://]REPO for container registry.")
	cmd.Flags().StringVarP(&cmdData.To, "to", "", "", "Destination address to copy stages to. Use archive:PATH for stage archive or [docker://]REPO for container registry.")
	cmd.Flags().BoolVarP(&cmdData.All, "all", "", true, "Copy all project stages (default: true). If false, copy only stages for current build.")

	return cmd
}

func runCopy(ctx context.Context) error {
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:                         &commonCmdData,
		InitWerf:                    true,
		InitGitDataManager:          true,
		InitDockerRegistry:          true,
		InitProcessContainerBackend: true,
		InitManifestCache:           true,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	if cmdData.From == "" {
		return fmt.Errorf("--from=ADDRESS param required")
	}

	if cmdData.To == "" {
		return fmt.Errorf("--to=ADDRESS param required")
	}

	if cmdData.From == cmdData.To {
		return fmt.Errorf("--from=ADDRESS and --to=ADDRESS must be different")
	}

	fromAddrRaw := cmdData.From
	toAddrRaw := cmdData.To

	fromAddr, err := ref.ParseAddr(fromAddrRaw)
	if err != nil {
		return fmt.Errorf("invalid from add %q: %w", fromAddrRaw, err)
	}

	toAddr, err := ref.ParseAddr(toAddrRaw)
	if err != nil {
		return fmt.Errorf("invalid to addr %q: %w", toAddrRaw, err)
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}
	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, config.WerfConfigOptions{LogRenderedFilePath: true, Env: commonCmdData.Environment})
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

	commonCmdData.Repo.Address = &fromAddr.RegistryAddress.Repo // FIXME выдумать что-нить симпатичнее

	dockerRegistry, err := common.CreateDockerRegistry(ctx, fromAddr.Repo, *commonCmdData.InsecureRegistry, *commonCmdData.SkipTlsVerifyRegistry)
	if err != nil {
		return fmt.Errorf("unable to create docker registry: %w", err)
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
		*fromAddr,
	)
	if err != nil {
		return fmt.Errorf("unable to init common copy components: %w", err)
	}

	var conveyorWithRetryWrapper *build.ConveyorWithRetryWrapper
	var buildOptions build.BuildOptions

	if !cmdData.All {
		conveyorWithRetryWrapper, buildOptions, err = initConveyorComponents(ctx, werfConfig, giterminismManager, projectTmpDir, containerBackend, storageManager)
		if err != nil {
			return fmt.Errorf("unable to init conveyor components: %w", err)
		}
		defer conveyorWithRetryWrapper.Terminate()
	}

	return logboek.Context(ctx).LogProcess("Copy stages").DoError(func() error {
		logboek.Context(ctx).LogFDetails("From: %s\n", fromAddr.String())
		logboek.Context(ctx).LogFDetails("To: %s\n", toAddr.String())

		return stages.Copy(ctx, fromAddr, toAddr, stages.CopyOptions{
			RegistryClient:    dockerRegistry,
			StorageManager:    storageManager,
			ConveyorWithRetry: conveyorWithRetryWrapper,
			All:               cmdData.All,
			ProjectName:       werfConfig.Meta.Project,
			BuildOptions:      buildOptions,
		})
	})
}

func initCommonCopyComponents(ctx context.Context, managerConfig *common.NewStorageManagerConfig, fromAddr ref.Addr) (*manager.StorageManager, docker_registry.Interface, error) {
	storageManager, err := common.NewStorageManager(ctx, managerConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init storage manager: %w", err)
	}

	dockerRegistry, err := common.CreateDockerRegistry(ctx, fromAddr.Repo, *commonCmdData.InsecureRegistry, *commonCmdData.SkipTlsVerifyRegistry)
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
