package cleanup

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/cleaning"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/stages_manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

var cmdData struct {
	SkipGitFetch              bool
	GitHistoryBasedCleanup    bool
	GitHistoryBasedCleanupV12 bool
}

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
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupStagesStorageOptions(&commonCmdData, cmd)
	common.SetupImagesRepoOptions(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to delete images from the specified images repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupImagesCleanupPolicies(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.SkipGitFetch, "skip-git-fetch", "", common.GetBoolEnvironmentDefaultFalse("WERF_SKIP_GIT_FETCH"), "Skip fetching and pruning unused git branches and tags (default $WERF_SKIP_GIT_FETCH)")
	cmd.Flags().BoolVarP(&cmdData.GitHistoryBasedCleanup, "git-history-based-cleanup", "", common.GetBoolEnvironmentDefaultFalse("WERF_GIT_HISTORY_BASED_CLEANUP"), "Use git history based cleanup (default $WERF_GIT_HISTORY_BASED_CLEANUP)")
	cmd.Flags().BoolVarP(&cmdData.GitHistoryBasedCleanupV12, "git-history-based-cleanup-v1.2", "", common.GetBoolEnvironmentDefaultFalse("WERF_GIT_HISTORY_BASED_CLEANUP_v1_2"), "Use git history based cleanup and delete images tags without related image metadata (default $WERF_GIT_HISTORY_BASED_CLEANUP_v1_2)")

	common.SetupDryRun(&commonCmdData, cmd)

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

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream(), LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
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

	logboek.LogOptionalLn()

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
		ImageNameList:             imagesNames,
		LocalGit:                  localRepo,
		KubernetesContextsClients: kubernetesContextsClients,
		WithoutKube:               *commonCmdData.WithoutKube,
		Policies:                  policies,
		SkipGitFetch:              cmdData.SkipGitFetch,
		GitHistoryBasedCleanup:    cmdData.GitHistoryBasedCleanup,
		GitHistoryBasedCleanupV12: cmdData.GitHistoryBasedCleanupV12,
		DryRun:                    *commonCmdData.DryRun,
	}

	logboek.LogOptionalLn()
	if err := cleaning.ImagesCleanup(projectName, imagesRepo, stagesManager, storageLockManager, imagesCleanupOptions); err != nil {
		return err
	}

	return nil
}
