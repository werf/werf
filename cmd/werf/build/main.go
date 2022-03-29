package build

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [IMAGE_NAME...]",
		Short: "Build images",
		Example: `  # Build images, built stages will be placed locally
  $ werf build

  # Build image 'backend'
  $ werf build backend

  # Build and enable drop-in shell session in the failed assembly container in the case when an error occurred
  $ werf build --introspect-error

  # Build images and store/use stages from repo
  $ werf build --repo harbor.company.io/werf`,
		Long: common.GetLongCommandDescription(`Build images that are described in werf.yaml.

The result of build command is built images pushed into the specified repo (or locally if repo is not specified).

If one or more IMAGE_NAME parameters specified, werf will build only these images`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := common.GetContext()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runMain(ctx, args)
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	common.SetupStagesStorageOptions(&commonCmdData, cmd)
	common.SetupFinalStagesStorageOptions(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupIntrospectAfterError(&commonCmdData, cmd)
	common.SetupIntrospectBeforeError(&commonCmdData, cmd)
	common.SetupIntrospectStage(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupReportPath(&commonCmdData, cmd)
	common.SetupReportFormat(&commonCmdData, cmd)

	common.SetupAddCustomTag(&commonCmdData, cmd)
	common.SetupVirtualMerge(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)
	common.SetupFollow(&commonCmdData, cmd)

	common.SetupDisableAutoHostCleanup(&commonCmdData, cmd)
	common.SetupAllowedDockerStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedDockerStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupDockerServerStoragePath(&commonCmdData, cmd)

	common.SetupPlatform(&commonCmdData, cmd)

	return cmd
}

func runMain(ctx context.Context, args []string) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	containerRuntime, processCtx, err := common.InitProcessContainerRuntime(ctx, &commonCmdData)
	if err != nil {
		return err
	}
	ctx = processCtx

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %s", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := lrumeta.Init(); err != nil {
		return err
	}

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
		return err
	}

	defer func() {
		if _, match := containerRuntime.(*container_runtime.DockerServerRuntime); !match {
			return
		}
		if err := common.RunAutoHostCleanup(ctx, &commonCmdData); err != nil {
			logboek.Context(ctx).Error().LogF("Auto host cleanup failed: %s\n", err)
		}
	}()

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	if err := ssh_agent.Init(ctx, common.GetSSHKey(&commonCmdData)); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	if *commonCmdData.Follow {
		logboek.LogOptionalLn()
		return common.FollowGitHead(ctx, &commonCmdData, func(ctx context.Context, headCommitGiterminismManager giterminism_manager.Interface) error {
			return run(ctx, containerRuntime, headCommitGiterminismManager, args)
		})
	} else {
		return run(ctx, containerRuntime, giterminismManager, args)
	}
}

func run(ctx context.Context, containerRuntime container_runtime.ContainerRuntime, giterminismManager giterminism_manager.Interface, imagesToProcess []string) error {
	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	if err := werfConfig.CheckThatImagesExist(imagesToProcess); err != nil {
		return err
	}

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	stagesStorageAddress := common.GetOptionalStagesStorageAddress(&commonCmdData)
	stagesStorage, err := common.GetStagesStorage(stagesStorageAddress, containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}
	finalStagesStorage, err := common.GetOptionalFinalStagesStorage(containerRuntime, &commonCmdData)
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
	secondaryStagesStorageList, err := common.GetSecondaryStagesStorageList(stagesStorage, containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}
	cacheStagesStorageList, err := common.GetCacheStagesStorageList(containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}

	storageManager := manager.NewStorageManager(projectName, stagesStorage, finalStagesStorage, secondaryStagesStorageList, cacheStagesStorageList, storageLockManager)

	buildOptions, err := common.GetBuildOptions(&commonCmdData, giterminismManager, werfConfig)
	if err != nil {
		return err
	}

	conveyorOptions, err := common.GetConveyorOptionsWithParallel(&commonCmdData, buildOptions)
	if err != nil {
		return err
	}

	logboek.LogOptionalLn()

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, imagesToProcess, giterminismManager.ProjectDir(), projectTmpDir, ssh_agent.SSHAuthSock, containerRuntime, storageManager, storageLockManager, conveyorOptions)
	defer conveyorWithRetry.Terminate()

	if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
		return c.Build(ctx, buildOptions)
	}); err != nil {
		return err
	}

	return nil
}
