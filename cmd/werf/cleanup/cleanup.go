package cleanup

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/cleaning"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/stages_manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
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
			defer werf.PrintGlobalWarnings(common.BackgroundContext())

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

	common.SetupGitHistorySynchronization(&commonCmdData, cmd)
	common.SetupAllowGitShallowClone(&commonCmdData, cmd)

	common.SetupScanContextNamespaceOnly(&commonCmdData, cmd)
	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)
	common.SetupWithoutKube(&commonCmdData, cmd)

	return cmd
}

func runCleanup() error {
	tmp_manager.AutoGCEnabled = true
	ctx := common.BackgroundContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := true_git.Init(true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(&commonCmdData); err != nil {
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

	if err := kube.Init(kube.InitOptions{KubeConfigOptions: kube.KubeConfigOptions{
		Context:          *commonCmdData.KubeContext,
		ConfigPath:       *commonCmdData.KubeConfig,
		ConfigDataBase64: *commonCmdData.KubeConfigBase64,
	}}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	if err := common.InitKubedog(ctx); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
	}

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&commonCmdData, projectDir)

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	werfConfig, err := common.GetRequiredWerfConfig(ctx, projectDir, &commonCmdData, true)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	containerRuntime := &container_runtime.LocalDockerServerRuntime{} // TODO

	stagesStorage, err := common.GetStagesStorage(containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}

	synchronization, err := common.GetSynchronization(ctx, &commonCmdData, projectName, stagesStorage)
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

	stagesManager := stages_manager.NewStagesManager(projectName, storageLockManager, stagesStorageCache)
	if err := stagesManager.UseStagesStorage(ctx, stagesStorage); err != nil {
		return err
	}

	imagesRepo, err := common.GetImagesRepo(ctx, projectName, &commonCmdData)
	if err != nil {
		return err
	}

	imagesNames, err := common.GetManagedImagesNames(ctx, projectName, stagesStorage, werfConfig)
	if err != nil {
		return err
	}
	logboek.Debug().LogF("Managed images names: %v\n", imagesNames)

	localGitRepo, err := common.GetLocalGitRepoForImagesCleanup(projectDir, &commonCmdData)
	if err != nil {
		return err
	}

	kubernetesContextClients, err := common.GetKubernetesContextClients(&commonCmdData)
	if err != nil {
		return fmt.Errorf("unable to get Kubernetes clusters connections: %s", err)
	}

	cleanupOptions := cleaning.CleanupOptions{
		ImagesCleanupOptions: cleaning.ImagesCleanupOptions{
			ImageNameList:                           imagesNames,
			LocalGit:                                localGitRepo,
			KubernetesContextClients:                kubernetesContextClients,
			KubernetesNamespaceRestrictionByContext: common.GetKubernetesNamespaceRestrictionByContext(&commonCmdData, kubernetesContextClients),
			WithoutKube:                             *commonCmdData.WithoutKube,
			GitHistoryBasedCleanupOptions:           werfConfig.Meta.Cleanup,
			DryRun:                                  *commonCmdData.DryRun,
		},
		StagesCleanupOptions: cleaning.StagesCleanupOptions{
			ImageNameList: imagesNames,
			DryRun:        *commonCmdData.DryRun,
		},
	}

	logboek.LogOptionalLn()
	if err := cleaning.Cleanup(ctx, projectName, imagesRepo, storageLockManager, stagesManager, cleanupOptions); err != nil {
		return err
	}

	return nil
}
