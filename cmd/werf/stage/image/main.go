package run

import (
	"fmt"
	"strings"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/werf/werf/pkg/image"

	"github.com/werf/werf/pkg/stages_manager"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "image [options] [IMAGE_NAME]",
		Short:                 "Print stage image name",
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Annotations: map[string]string{
			common.DisableOptionsInUseLineAnno: "1",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			//if err := common.ProcessLogOptions(&commonCmdData); err != nil {
			//	common.PrintHelp(cmd)
			//	return err
			//}
			logging.EnableLogQuiet()

			var imageName string
			if len(args) > 1 {
				common.PrintHelp(cmd)
				return fmt.Errorf("%d position argument can be specified, received %d", 1, len(args))
			} else if len(args) == 1 {
				imageName = args[0]
			}

			return run(imageName)
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupStagesStorageOptions(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and pull images from the specified stages storage")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogProjectDir(&commonCmdData, cmd)
	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupVirtualMerge(&commonCmdData, cmd)
	common.SetupVirtualMergeFromCommit(&commonCmdData, cmd)
	common.SetupVirtualMergeIntoCommit(&commonCmdData, cmd)

	return cmd
}

func run(imageName string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	tmp_manager.AutoGCEnabled = false

	if err := image.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream(), LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
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

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, &commonCmdData, false)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	if err := ssh_agent.Init(*commonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.LogWarnF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	if imageName == "" && len(werfConfig.StapelImages) == 1 {
		imageName = werfConfig.StapelImages[0].Name
	}

	if !werfConfig.HasImage(imageName) {
		return fmt.Errorf("image '%s' is not defined in werf.yaml", logging.ImageLogName(imageName, false))
	}

	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO

	stagesStorage, err := common.GetStagesStorage(containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}

	synchronization, err := common.GetSynchronization(&commonCmdData, stagesStorage.Address())
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
	if err := stagesManager.UseStagesStorage(stagesStorage); err != nil {
		return err
	}

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, []string{imageName}, projectDir, projectTmpDir, ssh_agent.SSHAuthSock, containerRuntime, stagesManager, nil, storageLockManager, common.GetConveyorOptions(&commonCmdData))
	defer conveyorWithRetry.Terminate()

	if err := conveyorWithRetry.WithRetryBlock(func(c *build.Conveyor) error {
		if err = c.ShouldBeBuilt(build.ShouldBeBuiltOptions{}); err != nil {
			return err
		}

		fmt.Println(c.GetImageNameForLastImageStage(imageName))

		return nil
	}); err != nil {
		return err
	}

	return nil
}
