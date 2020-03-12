package cleanup

import (
	"fmt"
	"path/filepath"

	"github.com/flant/werf/pkg/storage"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
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

	common.SetupStagesStorage(&commonCmdData, cmd)
	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupImagesRepo(&commonCmdData, cmd)
	common.SetupImagesRepoMode(&commonCmdData, cmd)
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

	if err := docker_registry.Init(docker_registry.Options{InsecureRegistry: *commonCmdData.InsecureRegistry, SkipTlsVerifyRegistry: *commonCmdData.SkipTlsVerifyRegistry}); err != nil {
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

	imagesRepoAddress, err := common.GetImagesRepoAddress(projectName, &commonCmdData)
	if err != nil {
		return err
	}
	imagesRepoMode, err := common.GetImagesRepoMode(&commonCmdData)
	if err != nil {
		return err
	}
	imagesRepoManager, err := storage.GetImagesRepoManager(imagesRepoAddress, imagesRepoMode)
	if err != nil {
		return err
	}
	imagesRepo := storage.NewDockerImagesRepo(projectName, imagesRepoManager)
	_ = imagesRepo // FIXME

	stagesStorageAddress, err := common.GetStagesStorageAddress(&commonCmdData)
	if err != nil {
		return err
	}
	_ = stagesStorageAddress // FIXME: parse stages storage address and create correct object
	stagesStorage := storage.NewLocalStagesStorage()
	stagesStorageCache := storage.NewFileStagesStorageCache(filepath.Join(werf.GetLocalCacheDir(), "stages_storage"))
	storageLockManager := &storage.FileLockManager{}
	_ = stagesStorageCache // FIXME
	_ = storageLockManager // FIXME

	imagesNames, err := common.GetManagedImagesNames(projectName, stagesStorage, werfConfig)
	if err != nil {
		return err
	}
	logboek.Debug.LogF("Managed images names: %v\n", imagesNames)

	stagesCleanupOptions := cleaning.StagesCleanupOptions{
		ProjectName:       projectName,
		ImagesRepoManager: imagesRepoManager, // FIXME: use imagesRepo only
		StagesStorage:     stagesStorage,
		ImagesNames:       imagesNames,
		DryRun:            *commonCmdData.DryRun,
	}

	logboek.LogOptionalLn()
	if err := cleaning.StagesCleanup(stagesCleanupOptions); err != nil {
		return err
	}

	return nil
}
