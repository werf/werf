package cleanup

import (
	"fmt"
	"path/filepath"

	"github.com/flant/werf/pkg/storage"

	"github.com/spf13/cobra"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "cleanup",
		DisableFlagsInUseLine: true,
		Short:                 "Cleanup project images from images repo",
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
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupStagesStorage(&commonCmdData, cmd)
	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupImagesRepo(&commonCmdData, cmd)
	common.SetupImagesRepoMode(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to delete images from the specified images repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupImagesCleanupPolicies(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupWithoutKube(&commonCmdData, cmd)

	return cmd
}

func runCleanup() error {
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

	var localRepo cleaning.GitRepo
	gitDir := filepath.Join(projectDir, ".git")
	if exist, err := util.DirExists(gitDir); err != nil {
		return err
	} else if exist {
		localRepo = &git_repo.Local{
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

	imagesCleanupOptions := cleaning.ImagesCleanupOptions{
		CommonRepoOptions: cleaning.CommonRepoOptions{
			ImagesRepoManager: imagesRepo.GetImagesRepoManager(),
			ImagesNames:       imagesNames,
			DryRun:            *commonCmdData.DryRun,
		},
		LocalGit:                  localRepo,
		KubernetesContextsClients: kubernetesContextsClients,
		WithoutKube:               *commonCmdData.WithoutKube,
		Policies:                  policies,
	}

	logboek.LogOptionalLn()
	if err := cleaning.ImagesCleanup(imagesCleanupOptions); err != nil {
		return err
	}

	return nil
}
