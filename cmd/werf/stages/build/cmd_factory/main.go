package cmd_factory

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

type CmdData struct {
	PullUsername string
	PullPassword string
}

func NewCmdWithData(cmdData *CmdData, commonCmdData *common.CmdData) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [IMAGE_NAME...]",
		Short: "Build stages",
		Example: `  # Build stages of all images from werf.yaml, built stages will be placed locally
  $ werf stages build --stages-storage :local

  # Build stages of image 'backend' from werf.yaml
  $ werf stages build --stages-storage :local backend

  # Build and enable drop-in shell session in the failed assembly container in the case when an error occurred
  $ werf build --stages-storage :local --introspect-error

  # Set --stages-storage default value using $WERF_STAGES_STORAGE param
  $ export WERF_STAGES_STORAGE=:local
  $ werf build`,
		Long: common.GetLongCommandDescription(`Build stages for images described in the werf.yaml.

The result of build command are built stages pushed into the specified stages storage (or locally in the case when --stages-storage=:local).

If one or more IMAGE_NAME parameters specified, werf will build only these images stages from werf.yaml`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			defer werf.PrintGlobalWarnings(common.BackgroundContext())

			if err := common.ProcessLogOptions(commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runStagesBuild(cmdData, commonCmdData, args)
			})
		},
	}

	common.SetupDir(commonCmdData, cmd)
	common.SetupConfigPath(commonCmdData, cmd)
	common.SetupConfigTemplatesDir(commonCmdData, cmd)
	common.SetupTmpDir(commonCmdData, cmd)
	common.SetupHomeDir(commonCmdData, cmd)
	common.SetupSSHKey(commonCmdData, cmd)

	common.SetupStagesStorageOptions(commonCmdData, cmd)

	common.SetupDockerConfig(commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified stages storage, to pull base images")
	common.SetupInsecureRegistry(commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(commonCmdData, cmd)

	common.SetupIntrospectAfterError(commonCmdData, cmd)
	common.SetupIntrospectBeforeError(commonCmdData, cmd)
	common.SetupIntrospectStage(commonCmdData, cmd)

	common.SetupLogOptions(commonCmdData, cmd)
	common.SetupLogProjectDir(commonCmdData, cmd)

	common.SetupSynchronization(commonCmdData, cmd)
	common.SetupKubeConfig(commonCmdData, cmd)
	common.SetupKubeConfigBase64(commonCmdData, cmd)
	common.SetupKubeContext(commonCmdData, cmd)

	common.SetupVirtualMerge(commonCmdData, cmd)
	common.SetupVirtualMergeFromCommit(commonCmdData, cmd)
	common.SetupVirtualMergeIntoCommit(commonCmdData, cmd)

	common.SetupGitUnshallow(commonCmdData, cmd)
	common.SetupAllowGitShallowClone(commonCmdData, cmd)
	common.SetupParallelOptions(commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	return cmd
}

func runStagesBuild(cmdData *CmdData, commonCmdData *common.CmdData, imagesToProcess []string) error {
	tmp_manager.AutoGCEnabled = true
	ctx := common.BackgroundContext()

	werf.PostponeMultiwerfNotUpToDateWarning()
	werf.PostponeWerf11DeprecationWarning()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := lrumeta.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(commonCmdData); err != nil {
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

	projectDir, err := common.GetProjectDir(commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(commonCmdData, projectDir)

	werfConfig, err := common.GetRequiredWerfConfig(ctx, projectDir, commonCmdData, true)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	for _, imageToProcess := range imagesToProcess {
		if !werfConfig.HasImageOrArtifact(imageToProcess) {
			return fmt.Errorf("specified image %s is not defined in werf.yaml", logging.ImageLogName(imageToProcess, false))
		}
	}

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO

	stagesStorage, err := common.GetStagesStorage(containerRuntime, commonCmdData)
	if err != nil {
		return err
	}

	synchronization, err := common.GetSynchronization(ctx, commonCmdData, projectName, stagesStorage)
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

	if err := ssh_agent.Init(ctx, common.GetSSHKey(commonCmdData)); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	buildStagesOptions, err := common.GetBuildStagesOptions(commonCmdData, werfConfig)
	if err != nil {
		return err
	}

	conveyorOptions, err := common.GetConveyorOptionsWithParallel(commonCmdData, buildStagesOptions)
	if err != nil {
		return err
	}

	logboek.LogOptionalLn()

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, imagesToProcess, projectDir, projectTmpDir, ssh_agent.SSHAuthSock, containerRuntime, storageManager, nil, storageLockManager, conveyorOptions)
	defer conveyorWithRetry.Terminate()

	if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
		return c.BuildStages(ctx, buildStagesOptions)
	}); err != nil {
		return err
	}

	return nil
}
