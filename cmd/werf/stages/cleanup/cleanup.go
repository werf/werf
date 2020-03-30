package cleanup

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/storage"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "cleanup",
		DisableFlagsInUseLine: true,
		Short:                 "Cleanup project stages from stages storage",
		Long:                  common.GetLongCommandDescription(`Cleanup project stages from stages storage for the images, that do not exist in the specified images repo`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runSync()
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupStagesStorageOptions(&commonCmdData, cmd)
	common.SetupImagesRepoOptions(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and delete images from the specified stages storage, read images from the specified images repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	return cmd
}

func runSync() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
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

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, true)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	logboek.LogOptionalLn()

	projectName := werfConfig.Meta.Project

	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO

	stagesStorage, err := common.GetStagesStorage(containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}

	stagesStorageCache := common.GetStagesStorageCache()
	_ = stagesStorageCache // FIXME

	storageLockManager := &storage.FileLockManager{}
	_ = storageLockManager // FIXME

	imagesRepo, err := common.GetImagesRepo(projectName, &commonCmdData)
	if err != nil {
		return err
	}

	imagesNames, err := common.GetManagedImagesNames(projectName, stagesStorage, werfConfig)
	if err != nil {
		return err
	}
	logboek.Debug.LogF("Managed images names: %v\n", imagesNames)

	stagesCleanupOptions := cleaning.StagesCleanupOptions{
		ImageNameList: imagesNames,
		DryRun:        *commonCmdData.DryRun,
	}

	logboek.LogOptionalLn()
	return cleaning.StagesCleanup(projectName, imagesRepo, stagesStorage, stagesCleanupOptions)
}
