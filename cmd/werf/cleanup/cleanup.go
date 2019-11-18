package cleanup

import (
	"fmt"
	"path/filepath"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/cleaning"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"

	"github.com/spf13/cobra"
)

var CommonCmdData common.CmdData

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
			if err := common.ProcessLogOptions(&CommonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}
			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runCleanup()
			})
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	common.SetupStagesStorage(&CommonCmdData, cmd)
	common.SetupImagesRepo(&CommonCmdData, cmd)
	common.SetupImagesRepoMode(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command needs granted permissions to read, pull and delete images from the specified stages storage and images repo")
	common.SetupInsecureRegistry(&CommonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&CommonCmdData, cmd)
	common.SetupImagesCleanupPolicies(&CommonCmdData, cmd)

	common.SetupKubeConfig(&CommonCmdData, cmd)
	common.SetupKubeContext(&CommonCmdData, cmd)

	common.SetupDryRun(&CommonCmdData, cmd)

	common.SetupLogOptions(&CommonCmdData, cmd)
	common.SetupLogProjectDir(&CommonCmdData, cmd)

	common.SetupWithoutKube(&CommonCmdData, cmd)

	return cmd
}

func runCleanup() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := docker_registry.Init(docker_registry.Options{InsecureRegistry: *CommonCmdData.InsecureRegistry, SkipTlsVerifyRegistry: *CommonCmdData.SkipTlsVerifyRegistry}); err != nil {
		return err
	}

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	if err := kube.Init(kube.InitOptions{KubeContext: *CommonCmdData.KubeContext, KubeConfig: *CommonCmdData.KubeConfig}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	if err := common.InitKubedog(); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(&CommonCmdData, projectDir)

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("bad config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	imagesRepo, err := common.GetImagesRepo(projectName, &CommonCmdData)
	if err != nil {
		return err
	}

	imagesRepoMode, err := common.GetImagesRepoMode(&CommonCmdData)
	if err != nil {
		return err
	}

	imagesRepoManager, err := common.GetImagesRepoManager(imagesRepo, imagesRepoMode)
	if err != nil {
		return err
	}

	stagesRepo, err := common.GetStagesRepo(&CommonCmdData)
	if err != nil {
		return err
	}

	var imagesNames []string
	for _, image := range werfConfig.StapelImages {
		imagesNames = append(imagesNames, image.Name)
	}

	for _, image := range werfConfig.ImagesFromDockerfile {
		imagesNames = append(imagesNames, image.Name)
	}

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

	policies, err := common.GetImagesCleanupPolicies(&CommonCmdData)
	if err != nil {
		return err
	}

	kubernetesContextsClients, err := kube.GetAllContextsClients(kube.GetAllContextsClientsOptions{KubeConfig: *CommonCmdData.KubeConfig})
	if err != nil {
		return fmt.Errorf("unable to get kubernetes clusters connections: %s", err)
	}

	imagesCleanupOptions := cleaning.ImagesCleanupOptions{
		CommonRepoOptions: cleaning.CommonRepoOptions{
			ImagesRepoManager: imagesRepoManager,
			ImagesNames:       imagesNames,
			DryRun:            *CommonCmdData.DryRun,
		},
		LocalGit:                  localGitRepo,
		KubernetesContextsClients: kubernetesContextsClients,
		WithoutKube:               *CommonCmdData.WithoutKube,
		Policies:                  policies,
	}

	stagesCleanupOptions := cleaning.StagesCleanupOptions{
		ProjectName:       projectName,
		ImagesRepoManager: imagesRepoManager,
		StagesStorage:     stagesRepo,
		ImagesNames:       imagesNames,
		DryRun:            *CommonCmdData.DryRun,
	}

	cleanupOptions := cleaning.CleanupOptions{
		StagesCleanupOptions: stagesCleanupOptions,
		ImagesCleanupOptions: imagesCleanupOptions,
	}

	logboek.LogOptionalLn()
	if err := cleaning.Cleanup(cleanupOptions); err != nil {
		return err
	}

	return nil
}
