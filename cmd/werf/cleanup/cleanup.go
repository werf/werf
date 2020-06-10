package cleanup

import (
	"fmt"
	"path/filepath"

	"github.com/werf/werf/pkg/image"

	"github.com/werf/werf/pkg/stages_manager"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/cleaning"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"

	"github.com/spf13/cobra"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "cleanup",
		DisableFlagsInUseLine: true,
		Short:                 "Safely cleanup unused project images and stages",
		Long: common.GetLongCommandDescription(`Safely cleanup unused project images and stages.

First step is 'werf images cleanup' command, which will delete unused images from images repo. Second step is 'werf stages cleanup' command, which will delete unused stages from stages storage to be in sync with the images repo.

It is safe to run this command periodically (daily is enough) by automated cleanup job in parallel with other werf commands such as build, deploy and host cleanup.`),
		Example: `  $ werf cleanup --stages-storage :local --images-repo registry.mydomain.com/myproject`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runCleanup()
			})
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupStagesStorageOptions(&commonCmdData, cmd)
	common.SetupImagesRepoOptions(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and delete images from the specified stages storage and images repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupImagesCleanupPolicies(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)
	common.SetupWithoutKube(&commonCmdData, cmd)

	return cmd
}

func runCleanup() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(&commonCmdData); err != nil {
		return err
	}

	if err := docker.Init(*commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	if err := kube.Init(kube.InitOptions{KubeContext: *commonCmdData.KubeContext, KubeConfig: *commonCmdData.KubeConfig}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	if err := common.InitKubedog(); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
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

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, &commonCmdData, true)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO

	stagesStorage, err := common.GetStagesStorage(containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}

	synchronization, err := common.GetSynchronization(&commonCmdData, stagesStorage.Address())
	if err != nil {
		return err
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

	imagesRepo, err := common.GetImagesRepo(projectName, &commonCmdData)
	if err != nil {
		return err
	}

	imagesNames, err := common.GetManagedImagesNames(projectName, stagesStorage, werfConfig)
	if err != nil {
		return err
	}
	logboek.Debug.LogF("Managed images names: %v\n", imagesNames)

	var localGitRepo cleaning.GitRepo
	gitDir := filepath.Join(projectDir, ".git")
	if exist, err := util.DirExists(gitDir); err != nil {
		return err
	} else if exist {
		localGitRepo = &git_repo.Local{
			Path:   projectDir,
			GitDir: gitDir,
		}
	}

	policies, err := common.GetImagesCleanupPolicies(&commonCmdData)
	if err != nil {
		return err
	}

	kubernetesContextsClients, err := kube.GetAllContextsClients(kube.GetAllContextsClientsOptions{KubeConfig: *commonCmdData.KubeConfig})
	if err != nil {
		return fmt.Errorf("unable to get Kubernetes clusters connections: %s", err)
	}

	cleanupOptions := cleaning.CleanupOptions{
		ImagesCleanupOptions: cleaning.ImagesCleanupOptions{
			ImageNameList:             imagesNames,
			LocalGit:                  localGitRepo,
			KubernetesContextsClients: kubernetesContextsClients,
			WithoutKube:               *commonCmdData.WithoutKube,
			Policies:                  policies,
			DryRun:                    *commonCmdData.DryRun,
		},
		StagesCleanupOptions: cleaning.StagesCleanupOptions{
			ImageNameList: imagesNames,
			DryRun:        *commonCmdData.DryRun,
		},
	}

	logboek.LogOptionalLn()
	if err := cleaning.Cleanup(projectName, imagesRepo, storageLockManager, stagesManager, cleanupOptions); err != nil {
		return err
	}

	return nil
}
