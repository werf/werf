package run

import (
	"fmt"
	"path/filepath"

	"github.com/flant/werf/pkg/container_runtime"

	"github.com/flant/werf/pkg/storage"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
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
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupStagesStorage(&commonCmdData, cmd)
	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and pull images from the specified stages storage")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogProjectDir(&commonCmdData, cmd)
	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	return cmd
}

func run(imageName string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	tmp_manager.AutoGCEnabled = false

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream(), LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := docker_registry.Init(*commonCmdData.InsecureRegistry, *commonCmdData.SkipTlsVerifyRegistry); err != nil {
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

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, false)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	_, err = common.GetStagesStorageAddress(&commonCmdData)
	if err != nil {
		return err
	}

	_, err = common.GetSynchronization(&commonCmdData)
	if err != nil {
		return err
	}

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

	stagesStorageAddress, err := common.GetStagesStorageAddress(&commonCmdData)
	if err != nil {
		return err
	}
	containerRuntime := &container_runtime.LocalDockerServerRuntime{}
	stagesStorage, err := storage.NewStagesStorage(stagesStorageAddress, containerRuntime)
	if err != nil {
		return err
	}

	stagesStorageCache := storage.NewFileStagesStorageCache(filepath.Join(werf.GetLocalCacheDir(), "stages_storage"))

	storageLockManager := &storage.FileLockManager{}

	c := build.NewConveyor(werfConfig, []string{imageName}, projectDir, projectTmpDir, ssh_agent.SSHAuthSock, containerRuntime, stagesStorage, stagesStorageCache, nil, storageLockManager)
	defer c.Terminate()

	if err = c.ShouldBeBuilt(); err != nil {
		return err
	}

	fmt.Println(c.GetImageNameForLastImageStage(imageName))

	return nil
}
