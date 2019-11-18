package cleanup

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

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
)

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "cleanup",
		DisableFlagsInUseLine: true,
		Short:                 "Cleanup project images from images repo",
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

	common.SetupImagesRepo(&CommonCmdData, cmd)
	common.SetupImagesRepoMode(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command needs granted permissions to delete images from the specified images repo")
	common.SetupInsecureRegistry(&CommonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&CommonCmdData, cmd)
	common.SetupImagesCleanupPolicies(&CommonCmdData, cmd)

	common.SetupKubeConfig(&CommonCmdData, cmd)
	common.SetupKubeContext(&CommonCmdData, cmd)

	common.SetupLogOptions(&CommonCmdData, cmd)
	common.SetupLogProjectDir(&CommonCmdData, cmd)

	common.SetupDryRun(&CommonCmdData, cmd)

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

	var imagesNames []string
	for _, image := range werfConfig.StapelImages {
		imagesNames = append(imagesNames, image.Name)
	}

	for _, image := range werfConfig.ImagesFromDockerfile {
		imagesNames = append(imagesNames, image.Name)
	}

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
		LocalGit:                  localRepo,
		KubernetesContextsClients: kubernetesContextsClients,
		WithoutKube:               *CommonCmdData.WithoutKube,
		Policies:                  policies,
	}

	logboek.LogOptionalLn()
	if err := cleaning.ImagesCleanup(imagesCleanupOptions); err != nil {
		return err
	}

	return nil
}
